package main

import (
	"context"
	"fmt"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

func cmdPlaceOrder(ctx context.Context, adp exchanges.Exchange, side exchanges.OrderSide, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "buy/sell <symbol> <quantity> [--price P] [--tif GTC|IOC|FOK|PO] [--post-only] [--reduce-only]"); err != nil {
		return err
	}

	symbol := args[0]
	qty, err := parseDecimal(args[1])
	if err != nil {
		return fmt.Errorf("invalid quantity: %s", args[1])
	}

	// Parse optional flags from remaining args
	remaining := args[2:]
	price, remaining := parsePrice(remaining)
	reduceOnly, remaining := parseBool(remaining, "--reduce-only")
	postOnly, remaining := parseBool(remaining, "--post-only")
	tifStr, remaining := parseStringFlag(remaining, "--tif")
	clientID, _ := parseStringFlag(remaining, "--client-id")

	// Determine order type
	orderType := exchanges.OrderTypeMarket
	if price.IsPositive() {
		orderType = exchanges.OrderTypeLimit
	}

	// Resolve TimeInForce
	tif := exchanges.TimeInForce("")
	if postOnly {
		tif = exchanges.TimeInForcePO
	} else if tifStr != "" {
		switch tifStr {
		case "GTC", "gtc":
			tif = exchanges.TimeInForceGTC
		case "IOC", "ioc":
			tif = exchanges.TimeInForceIOC
		case "FOK", "fok":
			tif = exchanges.TimeInForceFOK
		case "PO", "po":
			tif = exchanges.TimeInForcePO
		default:
			return fmt.Errorf("invalid time-in-force: %s (use GTC, IOC, FOK, or PO)", tifStr)
		}
	} else if orderType == exchanges.OrderTypeLimit {
		tif = exchanges.TimeInForceGTC
	}

	params := &exchanges.OrderParams{
		Symbol:      symbol,
		Side:        side,
		Type:        orderType,
		Quantity:    qty,
		Price:       price,
		TimeInForce: tif,
		ReduceOnly:  reduceOnly,
		ClientID:    clientID,
	}

	order, err := adp.PlaceOrder(ctx, params)
	if err != nil {
		return fmt.Errorf("place order: %w", err)
	}

	if jsonOut {
		outputJSON(order)
		return nil
	}

	outputSuccess("Order placed")
	outputTable(
		[]string{"OrderID", "Symbol", "Side", "Type", "Price", "Qty", "TIF", "Status"},
		[][]string{{
			order.OrderID,
			order.Symbol,
			colorSide(string(order.Side)),
			string(order.Type),
			priceStr(order.Price),
			decStr(order.Quantity),
			valOrDash(string(order.TimeInForce)),
			colorStatus(string(order.Status)),
		}},
	)
	return nil
}

func cmdCancelOrder(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "cancel <orderID> <symbol>"); err != nil {
		return err
	}
	orderID := args[0]
	symbol := args[1]

	err := adp.CancelOrder(ctx, orderID, symbol)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	if jsonOut {
		outputJSON(map[string]string{"status": "cancelled", "orderID": orderID})
		return nil
	}

	outputSuccess("Order %s cancelled", orderID)
	return nil
}

func cmdCancelAllOrders(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "cancel-all <symbol>"); err != nil {
		return err
	}
	symbol := args[0]

	err := adp.CancelAllOrders(ctx, symbol)
	if err != nil {
		return fmt.Errorf("cancel all orders: %w", err)
	}

	if jsonOut {
		outputJSON(map[string]string{"status": "all_cancelled", "symbol": symbol})
		return nil
	}

	outputSuccess("All orders for %s cancelled", symbol)
	return nil
}

func cmdModifyOrder(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "modify <orderID> <symbol> [--price P] [--qty Q]"); err != nil {
		return err
	}

	perpAdp, ok := adp.(exchanges.PerpExchange)
	if !ok {
		return fmt.Errorf("modify order is only available for perpetual contracts (current: %s mode)", adp.GetMarketType())
	}

	orderID := args[0]
	symbol := args[1]

	remaining := args[2:]
	price, remaining := parsePrice(remaining)
	qty, _ := parseQty(remaining)

	params := &exchanges.ModifyOrderParams{
		Price:    price,
		Quantity: qty,
	}

	order, err := perpAdp.ModifyOrder(ctx, orderID, symbol, params)
	if err != nil {
		return fmt.Errorf("modify order: %w", err)
	}

	if jsonOut {
		outputJSON(order)
		return nil
	}

	outputSuccess("Order modified")
	outputTable(
		[]string{"OrderID", "Symbol", "Side", "Type", "Price", "Qty", "Status"},
		[][]string{{
			order.OrderID,
			order.Symbol,
			colorSide(string(order.Side)),
			string(order.Type),
			priceStr(order.Price),
			decStr(order.Quantity),
			colorStatus(string(order.Status)),
		}},
	)
	return nil
}

func cmdFetchOrder(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "order <orderID> <symbol>"); err != nil {
		return err
	}

	orderID := args[0]
	symbol := args[1]

	order, err := adp.FetchOrderByID(ctx, orderID, symbol)
	if err != nil {
		return fmt.Errorf("fetch order: %w", err)
	}

	if jsonOut {
		outputJSON(order)
		return nil
	}

	outputTable(
		[]string{"OrderID", "Symbol", "Side", "Type", "Price", "Qty", "Filled", "Status"},
		[][]string{{
			order.OrderID,
			order.Symbol,
			colorSide(string(order.Side)),
			string(order.Type),
			priceStr(displayOrderPrice(order)),
			decStr(order.Quantity),
			decStr(order.FilledQuantity),
			colorStatus(string(order.Status)),
		}},
	)
	return nil
}

func priceStr(p decimal.Decimal) string {
	if p.IsZero() {
		return "-"
	}
	return p.String()
}

func displayOrderPrice(order *exchanges.Order) decimal.Decimal {
	if order == nil {
		return decimal.Zero
	}
	switch {
	case !order.OrderPrice.IsZero():
		return order.OrderPrice
	case !order.AverageFillPrice.IsZero():
		return order.AverageFillPrice
	case !order.LastFillPrice.IsZero():
		return order.LastFillPrice
	default:
		return order.Price
	}
}
