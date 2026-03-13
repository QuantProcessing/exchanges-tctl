package main

import (
	"context"
	"fmt"
	"strconv"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

// dispatch routes the CLI command to the appropriate handler.
func dispatch(ctx context.Context, adp exchanges.Exchange, exchName, cmd string, args []string, jsonOut bool) error {
	switch cmd {
	// Market data
	case "ticker", "t":
		return cmdTicker(ctx, adp, args, jsonOut)
	case "orderbook", "ob":
		return cmdOrderBook(ctx, adp, args, jsonOut)
	case "details":
		return cmdSymbolDetails(ctx, adp, args, jsonOut)
	case "fee":
		return cmdFeeRate(ctx, adp, args, jsonOut)
	case "funding":
		return cmdFundingRate(ctx, adp, args, jsonOut)
	case "trades":
		return cmdTrades(ctx, adp, args, jsonOut)
	case "klines", "kl":
		return cmdKlines(ctx, adp, args, jsonOut)

	// Trading
	case "buy":
		return cmdPlaceOrder(ctx, adp, exchanges.OrderSideBuy, args, jsonOut)
	case "sell":
		return cmdPlaceOrder(ctx, adp, exchanges.OrderSideSell, args, jsonOut)
	case "cancel":
		return cmdCancelOrder(ctx, adp, args, jsonOut)
	case "cancel-all":
		return cmdCancelAllOrders(ctx, adp, args, jsonOut)
	case "modify":
		return cmdModifyOrder(ctx, adp, args, jsonOut)
	case "order":
		return cmdFetchOrder(ctx, adp, args, jsonOut)

	// Account (common)
	case "orders", "o":
		return cmdOpenOrders(ctx, adp, args, jsonOut)
	case "balance", "bal", "b":
		return cmdBalance(ctx, adp, jsonOut)
	case "account", "acc":
		return cmdAccount(ctx, adp, jsonOut)

	// Perp-specific
	case "positions", "pos", "p":
		return cmdPositions(ctx, adp, args, jsonOut)
	case "leverage", "lev":
		return cmdSetLeverage(ctx, adp, args, jsonOut)
	case "funding-all":
		return cmdAllFundingRates(ctx, adp, jsonOut)

	// Spot-specific
	case "spot-balances", "sb":
		return cmdSpotBalances(ctx, adp, jsonOut)
	case "transfer":
		return cmdTransfer(ctx, adp, args, jsonOut)

	// WSS streaming
	case "watch-ticker", "wt":
		return cmdWatchTicker(ctx, adp, args, jsonOut)
	case "watch-ob", "wob":
		return cmdWatchOrderBook(ctx, adp, args, jsonOut)
	case "watch-orders", "wo":
		return cmdWatchOrders(ctx, adp, args, jsonOut)
	case "watch-trades", "wtr":
		return cmdWatchTrades(ctx, adp, args, jsonOut)
	case "watch-positions", "wp":
		return cmdWatchPositions(ctx, adp, args, jsonOut)
	case "watch-klines", "wkl":
		return cmdWatchKlines(ctx, adp, args, jsonOut)

	default:
		return fmt.Errorf("unknown command: %s. Run 'tctl --help' for usage", cmd)
	}
}

// requireArgs checks that enough arguments were provided.
func requireArgs(args []string, min int, usage string) error {
	if len(args) < min {
		return fmt.Errorf("not enough arguments. Usage: %s", usage)
	}
	return nil
}

// parsePrice extracts --price from args.
func parsePrice(args []string) (decimal.Decimal, []string) {
	for i, a := range args {
		if a == "--price" && i+1 < len(args) {
			p, err := decimal.NewFromString(args[i+1])
			if err == nil {
				// Remove --price and value from args
				remaining := make([]string, 0, len(args)-2)
				remaining = append(remaining, args[:i]...)
				remaining = append(remaining, args[i+2:]...)
				return p, remaining
			}
		}
	}
	return decimal.Zero, args
}

// parseBool extracts a boolean flag from args.
func parseBool(args []string, flag string) (bool, []string) {
	for i, a := range args {
		if a == flag {
			remaining := make([]string, 0, len(args)-1)
			remaining = append(remaining, args[:i]...)
			remaining = append(remaining, args[i+1:]...)
			return true, remaining
		}
	}
	return false, args
}

// parseDecimal parses a decimal from string.
func parseDecimal(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

// parseInt parses an int from string.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseQty extracts --qty from args.
func parseQty(args []string) (decimal.Decimal, []string) {
	for i, a := range args {
		if a == "--qty" && i+1 < len(args) {
			q, err := decimal.NewFromString(args[i+1])
			if err == nil {
				remaining := make([]string, 0, len(args)-2)
				remaining = append(remaining, args[:i]...)
				remaining = append(remaining, args[i+2:]...)
				return q, remaining
			}
		}
	}
	return decimal.Zero, args
}

// parseStringFlag extracts a string flag value from args (e.g., --from SPOT).
func parseStringFlag(args []string, flag string) (string, []string) {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			val := args[i+1]
			remaining := make([]string, 0, len(args)-2)
			remaining = append(remaining, args[:i]...)
			remaining = append(remaining, args[i+2:]...)
			return val, remaining
		}
	}
	return "", args
}
