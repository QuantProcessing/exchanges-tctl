package main

import (
	"context"
	"fmt"
	"strings"

	exchanges "github.com/QuantProcessing/exchanges"
)

func cmdPositions(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	perpAdp, ok := adp.(exchanges.PerpExchange)
	if !ok {
		return fmt.Errorf("positions are only available for perpetual contracts (current: %s mode)", adp.GetMarketType())
	}

	positions, err := perpAdp.FetchPositions(ctx)
	if err != nil {
		return fmt.Errorf("fetch positions: %w", err)
	}

	if jsonOut {
		outputJSON(positions)
		return nil
	}

	if len(positions) == 0 {
		outputInfo("No open positions")
		return nil
	}

	rows := make([][]string, 0, len(positions))
	for _, p := range positions {
		rows = append(rows, []string{
			p.Symbol,
			colorSide(string(p.Side)),
			decStr(p.Quantity),
			decStr(p.EntryPrice),
			colorPnL(p.UnrealizedPnL),
			decStr(p.Leverage),
		})
	}
	outputTable([]string{"Symbol", "Side", "Qty", "Entry", "UnPnL", "Leverage"}, rows)
	return nil
}

func cmdOpenOrders(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	symbol := ""
	if len(args) > 0 {
		symbol = args[0]
	}

	orders, err := adp.FetchOpenOrders(ctx, symbol)
	if err != nil {
		return fmt.Errorf("fetch open orders: %w", err)
	}

	if jsonOut {
		outputJSON(orders)
		return nil
	}

	if len(orders) == 0 {
		outputInfo("No open orders")
		return nil
	}

	rows := make([][]string, 0, len(orders))
	for _, o := range orders {
		rows = append(rows, []string{
			o.OrderID,
			o.Symbol,
			colorSide(string(o.Side)),
			string(o.Type),
			priceStr(displayOrderPrice(&o)),
			decStr(o.Quantity),
			colorStatus(string(o.Status)),
		})
	}
	outputTable([]string{"OrderID", "Symbol", "Side", "Type", "Price", "Qty", "Status"}, rows)
	return nil
}

func cmdBalance(ctx context.Context, adp exchanges.Exchange, jsonOut bool) error {
	balance, err := adp.FetchBalance(ctx)
	if err != nil {
		return fmt.Errorf("fetch balance: %w", err)
	}

	if jsonOut {
		outputJSON(map[string]string{"balance": balance.String()})
		return nil
	}

	fmt.Printf("Balance: %s\n", colorBold(balance.String()))
	return nil
}

func cmdAccount(ctx context.Context, adp exchanges.Exchange, jsonOut bool) error {
	account, err := adp.FetchAccount(ctx)
	if err != nil {
		return fmt.Errorf("fetch account: %w", err)
	}

	if jsonOut {
		outputJSON(account)
		return nil
	}

	fmt.Println(colorBold("=== Balance ==="))
	outputTable(
		[]string{"Total", "Available"},
		[][]string{{
			decStr(account.TotalBalance),
			decStr(account.AvailableBalance),
		}},
	)

	if account.UnrealizedPnL.IsPositive() || account.UnrealizedPnL.IsNegative() {
		fmt.Printf("Unrealized PnL: %s\n", colorPnL(account.UnrealizedPnL))
	}

	if len(account.Positions) > 0 {
		fmt.Println("\n" + colorBold("=== Positions ==="))
		rows := make([][]string, 0, len(account.Positions))
		for _, p := range account.Positions {
			rows = append(rows, []string{
				p.Symbol,
				colorSide(string(p.Side)),
				decStr(p.Quantity),
				decStr(p.EntryPrice),
				colorPnL(p.UnrealizedPnL),
			})
		}
		outputTable([]string{"Symbol", "Side", "Qty", "Entry", "UnPnL"}, rows)
	}
	return nil
}

func cmdSetLeverage(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "leverage <symbol> <value>"); err != nil {
		return err
	}

	perpAdp, ok := adp.(exchanges.PerpExchange)
	if !ok {
		return fmt.Errorf("leverage is only available for perpetual contracts (current: %s mode)", adp.GetMarketType())
	}

	symbol := args[0]
	lev, err := parseInt(args[1])
	if err != nil {
		return fmt.Errorf("invalid leverage: %s", args[1])
	}

	err = perpAdp.SetLeverage(ctx, symbol, lev)
	if err != nil {
		return fmt.Errorf("set leverage: %w", err)
	}

	if jsonOut {
		outputJSON(map[string]interface{}{"status": "ok", "symbol": symbol, "leverage": lev})
		return nil
	}

	outputSuccess("Leverage for %s set to %dx", symbol, lev)
	return nil
}

// ============================================================================
// Spot-specific commands
// ============================================================================

func cmdSpotBalances(ctx context.Context, adp exchanges.Exchange, jsonOut bool) error {
	spotAdp, ok := adp.(exchanges.SpotExchange)
	if !ok {
		return fmt.Errorf("spot-balances is only available for spot market (current: %s mode)", adp.GetMarketType())
	}

	balances, err := spotAdp.FetchSpotBalances(ctx)
	if err != nil {
		return fmt.Errorf("fetch spot balances: %w", err)
	}

	if jsonOut {
		outputJSON(balances)
		return nil
	}

	if len(balances) == 0 {
		outputInfo("No spot balances")
		return nil
	}

	rows := make([][]string, 0, len(balances))
	for _, b := range balances {
		rows = append(rows, []string{
			b.Asset,
			decStr(b.Free),
			decStr(b.Locked),
			decStr(b.Total),
		})
	}
	outputTable([]string{"Asset", "Free", "Locked", "Total"}, rows)
	return nil
}

func cmdTransfer(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "transfer <asset> <amount> [--from SPOT] [--to PERP]"); err != nil {
		return err
	}

	spotAdp, ok := adp.(exchanges.SpotExchange)
	if !ok {
		return fmt.Errorf("transfer is only available for spot market (current: %s mode)", adp.GetMarketType())
	}

	asset := strings.ToUpper(args[0])
	amount, err := parseDecimal(args[1])
	if err != nil {
		return fmt.Errorf("invalid amount: %s", args[1])
	}

	remaining := args[2:]
	fromStr, remaining := parseStringFlag(remaining, "--from")
	toStr, _ := parseStringFlag(remaining, "--to")

	// Defaults
	fromAcct := exchanges.AccountTypeSpot
	toAcct := exchanges.AccountTypePerp

	if fromStr != "" {
		fromAcct = exchanges.AccountType(strings.ToUpper(fromStr))
	}
	if toStr != "" {
		toAcct = exchanges.AccountType(strings.ToUpper(toStr))
	}

	params := &exchanges.TransferParams{
		Asset:       asset,
		Amount:      amount,
		FromAccount: fromAcct,
		ToAccount:   toAcct,
	}

	err = spotAdp.TransferAsset(ctx, params)
	if err != nil {
		return fmt.Errorf("transfer asset: %w", err)
	}

	if jsonOut {
		outputJSON(map[string]interface{}{
			"status": "ok",
			"asset":  asset,
			"amount": amount.String(),
			"from":   string(fromAcct),
			"to":     string(toAcct),
		})
		return nil
	}

	outputSuccess("Transferred %s %s from %s to %s", amount.String(), asset, fromAcct, toAcct)
	return nil
}
