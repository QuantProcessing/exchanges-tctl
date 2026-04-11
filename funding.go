package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// cmdFundArb implements the funding rate arbitrage command.
// By default it creates both a spot and perp adapter for the same exchange.
// When both --spot-exchange and --perp-exchange are provided, each leg can use
// a different exchange while still sharing the same symbol and quantity.
func cmdFundArb(ctx context.Context, defaultExchange string, args []string, jsonOut bool, logger *zap.SugaredLogger) error {
	if err := requireArgs(args, 2, "fund-arb <symbol> <quantity> [--leverage N] [--close] [--spot-price P] [--perp-price P] [--spot-exchange EX] [--perp-exchange EX]"); err != nil {
		return err
	}

	symbol := args[0]
	qty, err := parseDecimal(args[1])
	if err != nil {
		return fmt.Errorf("invalid quantity: %s", args[1])
	}

	// Parse flags
	remaining := args[2:]
	closeMode, remaining := parseBool(remaining, "--close")
	spotPrice, remaining := parseNamedPrice(remaining, "--spot-price")
	perpPrice, remaining := parseNamedPrice(remaining, "--perp-price")
	spotExchange, remaining := parseStringFlag(remaining, "--spot-exchange")
	perpExchange, remaining := parseStringFlag(remaining, "--perp-exchange")
	leverageStr, _ := parseStringFlag(remaining, "--leverage")

	spotExchange, perpExchange, err = resolveFundArbExchanges(defaultExchange, spotExchange, perpExchange)
	if err != nil {
		return err
	}

	leverage := 1
	if leverageStr != "" {
		lev, err := parseInt(leverageStr)
		if err != nil || lev < 1 {
			return fmt.Errorf("invalid leverage: %s", leverageStr)
		}
		leverage = lev
	}

	// Create both adapters
	spotAdp, err := createRESTAdapter(ctx, spotExchange, exchanges.MarketTypeSpot, logger)
	if err != nil {
		return fmt.Errorf("failed to create spot adapter: %w", err)
	}
	perpAdp, err := createRESTAdapter(ctx, perpExchange, exchanges.MarketTypePerp, logger)
	if err != nil {
		return fmt.Errorf("failed to create perp adapter: %w", err)
	}

	// Close adapters on exit
	defer func() {
		if c, ok := spotAdp.(interface{ Close() error }); ok {
			c.Close()
		}
		if c, ok := perpAdp.(interface{ Close() error }); ok {
			c.Close()
		}
	}()

	// Determine sides
	var spotSide, perpSide exchanges.OrderSide
	if closeMode {
		spotSide = exchanges.OrderSideSell
		perpSide = exchanges.OrderSideBuy
	} else {
		spotSide = exchanges.OrderSideBuy
		perpSide = exchanges.OrderSideSell
	}

	// Set leverage (only when opening)
	if !closeMode && leverage > 1 {
		if perp, ok := perpAdp.(exchanges.PerpExchange); ok {
			if !jsonOut {
				outputInfo("Setting %s leverage to %dx...", symbol, leverage)
			}
			if err := perp.SetLeverage(ctx, symbol, leverage); err != nil {
				return fmt.Errorf("failed to set leverage: %w", err)
			}
		} else {
			return fmt.Errorf("%s does not support leverage setting", perpExchange)
		}
	}

	// Build order params
	spotParams := &exchanges.OrderParams{
		Symbol:   symbol,
		Side:     spotSide,
		Quantity: qty,
		Type:     exchanges.OrderTypeMarket,
	}
	if spotPrice.IsPositive() {
		spotParams.Type = exchanges.OrderTypeLimit
		spotParams.Price = spotPrice
	}

	perpParams := &exchanges.OrderParams{
		Symbol:   symbol,
		Side:     perpSide,
		Quantity: qty,
		Type:     exchanges.OrderTypeMarket,
	}
	if perpPrice.IsPositive() {
		perpParams.Type = exchanges.OrderTypeLimit
		perpParams.Price = perpPrice
	}
	if closeMode {
		perpParams.ReduceOnly = true
	}

	// Place orders concurrently
	if !jsonOut {
		action := "Opening"
		if closeMode {
			action = "Closing"
		}
		if spotExchange == perpExchange {
			outputInfo("%s funding arb: %s %s × %s", action, symbol, qty.String(), spotExchange)
		} else {
			outputInfo("%s funding arb: %s %s × spot=%s perp=%s", action, symbol, qty.String(), spotExchange, perpExchange)
		}
	}

	var (
		spotOrder, perpOrder *exchanges.Order
		spotErr, perpErr     error
		wg                   sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		spotOrder, spotErr = spotAdp.PlaceOrder(ctx, spotParams)
	}()
	go func() {
		defer wg.Done()
		perpOrder, perpErr = perpAdp.PlaceOrder(ctx, perpParams)
	}()
	wg.Wait()

	// Output results
	if jsonOut {
		return outputFundArbJSON(spotExchange, spotOrder, spotErr, perpExchange, perpOrder, perpErr)
	}
	return outputFundArbTable(spotExchange, spotOrder, spotErr, perpExchange, perpOrder, perpErr, closeMode)
}

func resolveFundArbExchanges(defaultExchange, spotExchange, perpExchange string) (string, string, error) {
	spotExchange = strings.ToUpper(strings.TrimSpace(spotExchange))
	perpExchange = strings.ToUpper(strings.TrimSpace(perpExchange))

	if (spotExchange == "") != (perpExchange == "") {
		return "", "", fmt.Errorf("must set both --spot-exchange and --perp-exchange")
	}
	if spotExchange != "" {
		return spotExchange, perpExchange, nil
	}

	resolved := resolveExchange(defaultExchange)
	if resolved == "" {
		return "", "", fmt.Errorf("no exchange specified and none auto-detected. Use -e flag or set both --spot-exchange and --perp-exchange")
	}
	return resolved, resolved, nil
}

// parseNamedPrice extracts a named price flag (e.g., --spot-price 95000).
func parseNamedPrice(args []string, flag string) (decimal.Decimal, []string) {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			p, err := decimal.NewFromString(args[i+1])
			if err == nil {
				remaining := make([]string, 0, len(args)-2)
				remaining = append(remaining, args[:i]...)
				remaining = append(remaining, args[i+2:]...)
				return p, remaining
			}
		}
	}
	return decimal.Zero, args
}

// outputFundArbJSON outputs both orders as JSON.
func outputFundArbJSON(spotExchange string, spotOrder *exchanges.Order, spotErr error, perpExchange string, perpOrder *exchanges.Order, perpErr error) error {
	type orderResult struct {
		Market   string           `json:"market"`
		Exchange string           `json:"exchange"`
		Order    *exchanges.Order `json:"order,omitempty"`
		Error    string           `json:"error,omitempty"`
	}

	results := make([]orderResult, 2)
	results[0] = orderResult{Market: "spot", Exchange: spotExchange}
	if spotErr != nil {
		results[0].Error = spotErr.Error()
	} else {
		results[0].Order = spotOrder
	}
	results[1] = orderResult{Market: "perp", Exchange: perpExchange}
	if perpErr != nil {
		results[1].Error = perpErr.Error()
	} else {
		results[1].Order = perpOrder
	}

	outputJSON(results)

	if spotErr != nil && perpErr != nil {
		return fmt.Errorf("both orders failed: spot=%v, perp=%v", spotErr, perpErr)
	}
	return nil
}

// outputFundArbTable outputs both orders as a formatted table.
func outputFundArbTable(spotExchange string, spotOrder *exchanges.Order, spotErr error, perpExchange string, perpOrder *exchanges.Order, perpErr error, closeMode bool) error {
	action := "OPEN"
	if closeMode {
		action = "CLOSE"
	}

	fmt.Println()
	fmt.Printf("  Funding Arb %s Results\n", action)
	fmt.Println("  " + strings.Repeat("─", 50))

	// Spot result
	if spotErr != nil {
		outputError("SPOT  (%s) ✗ %v", spotExchange, spotErr)
	} else {
		fmt.Printf("  SPOT  (%s) ✓ %s %s %s @ %s  [%s]\n",
			spotExchange,
			colorSide(string(spotOrder.Side)),
			spotOrder.Symbol,
			decStr(spotOrder.Quantity),
			priceStr(displayOrderPrice(spotOrder)),
			spotOrder.OrderID,
		)
	}

	// Perp result
	if perpErr != nil {
		outputError("PERP  (%s) ✗ %v", perpExchange, perpErr)
	} else {
		fmt.Printf("  PERP  (%s) ✓ %s %s %s @ %s  [%s]\n",
			perpExchange,
			colorSide(string(perpOrder.Side)),
			perpOrder.Symbol,
			decStr(perpOrder.Quantity),
			priceStr(displayOrderPrice(perpOrder)),
			perpOrder.OrderID,
		)
	}

	fmt.Println()

	// Warnings
	if spotErr != nil && perpErr == nil {
		outputError("⚠ SPOT order failed but PERP succeeded. Manual intervention needed!")
		return fmt.Errorf("spot order failed: %w", spotErr)
	}
	if perpErr != nil && spotErr == nil {
		outputError("⚠ PERP order failed but SPOT succeeded. Manual intervention needed!")
		return fmt.Errorf("perp order failed: %w", perpErr)
	}
	if spotErr != nil && perpErr != nil {
		return fmt.Errorf("both orders failed: spot=%v, perp=%v", spotErr, perpErr)
	}

	outputSuccess("Both orders placed successfully")
	return nil
}
