package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

func cmdWatchTicker(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "watch-ticker <symbol>"); err != nil {
		return err
	}
	symbol := args[0]

	fmt.Printf("Watching ticker for %s (Ctrl+C to stop)...\n\n", symbol)

	var lastPrice decimal.Decimal

	err := adp.WatchTicker(ctx, symbol, func(t *exchanges.Ticker) {
		if jsonOut {
			outputJSON(t)
			return
		}

		spread := t.Ask.Sub(t.Bid)

		// Delta indicator
		delta := ""
		if !lastPrice.IsZero() {
			cmp := t.LastPrice.Cmp(lastPrice)
			if cmp > 0 {
				delta = colorGreen("▲")
			} else if cmp < 0 {
				delta = colorRed("▼")
			} else {
				delta = colorDim("─")
			}
		}
		lastPrice = t.LastPrice

		fmt.Printf("\r\033[K%s %s Bid: %s  Ask: %s  Spread: %s  Last: %s  Vol: %s",
			colorCyan(t.Symbol),
			delta,
			colorGreen(decStr(t.Bid)),
			colorRed(decStr(t.Ask)),
			decStr(spread),
			decStr(t.LastPrice),
			decStr(t.Volume24h),
		)
	})
	if err != nil {
		return fmt.Errorf("watch ticker: %w", err)
	}

	// Block until context is cancelled
	<-ctx.Done()
	fmt.Println("\nStopped watching ticker.")
	return nil
}

func cmdWatchOrderBook(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "watch-ob <symbol> [depth]"); err != nil {
		return err
	}
	symbol := args[0]
	depth := 5
	if len(args) > 1 {
		if d, err := parseInt(args[1]); err == nil {
			depth = d
		}
	}

	fmt.Printf("Watching orderbook for %s (depth=%d, Ctrl+C to stop)...\n", symbol, depth)

	err := adp.WatchOrderBook(ctx, symbol, func(ob *exchanges.OrderBook) {
		if jsonOut {
			trimmed := trimOrderBook(ob, depth)
			outputJSON(trimmed)
			return
		}
		printLiveOrderBook(ob, depth)
	})
	if err != nil {
		return fmt.Errorf("watch orderbook: %w", err)
	}

	<-ctx.Done()
	fmt.Println("\nStopped watching orderbook.")
	return nil
}

func cmdWatchOrders(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	fmt.Println("Watching order updates (Ctrl+C to stop)...")

	err := adp.WatchOrders(ctx, func(o *exchanges.Order) {
		if jsonOut {
			outputJSON(o)
			return
		}
		ts := time.Now().Format("15:04:05")
		sideColor := colorGreen
		if o.Side == exchanges.OrderSideSell {
			sideColor = colorRed
		}
		fmt.Printf("[%s] %s %s %s  Price: %s  Qty: %s  Filled: %s  Status: %s\n",
			ts,
			sideColor(string(o.Side)),
			o.Symbol,
			string(o.Type),
			priceStr(o.Price),
			decStr(o.Quantity),
			decStr(o.FilledQuantity),
			colorStatus(string(o.Status)),
		)
	})
	if err != nil {
		return fmt.Errorf("watch orders: %w", err)
	}

	<-ctx.Done()
	fmt.Println("\nStopped watching orders.")
	return nil
}

func cmdWatchTrades(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "watch-trades <symbol>"); err != nil {
		return err
	}
	symbol := args[0]

	fmt.Printf("Watching trades for %s (Ctrl+C to stop)...\n", symbol)

	err := adp.WatchTrades(ctx, symbol, func(t *exchanges.Trade) {
		if jsonOut {
			outputJSON(t)
			return
		}
		ts := time.Now().Format("15:04:05")
		sideColor := colorGreen
		if t.Side == exchanges.TradeSideSell {
			sideColor = colorRed
		}
		fmt.Printf("[%s] %s  %s  Price: %s  Qty: %s\n",
			ts,
			sideColor(string(t.Side)),
			t.Symbol,
			decStr(t.Price),
			decStr(t.Quantity),
		)
	})
	if err != nil {
		return fmt.Errorf("watch trades: %w", err)
	}

	<-ctx.Done()
	fmt.Println("\nStopped watching trades.")
	return nil
}

func cmdWatchPositions(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	fmt.Println("Watching position updates (Ctrl+C to stop)...")

	err := adp.WatchPositions(ctx, func(p *exchanges.Position) {
		if jsonOut {
			outputJSON(p)
			return
		}
		ts := time.Now().Format("15:04:05")
		sideColor := colorGreen
		if p.Side == exchanges.PositionSideShort {
			sideColor = colorRed
		}
		fmt.Printf("[%s] %s %s  Qty: %s  Entry: %s  PnL: %s  Lev: %sx\n",
			ts,
			sideColor(string(p.Side)),
			p.Symbol,
			decStr(p.Quantity),
			decStr(p.EntryPrice),
			colorPnL(p.UnrealizedPnL),
			decStr(p.Leverage),
		)
	})
	if err != nil {
		return fmt.Errorf("watch positions: %w", err)
	}

	<-ctx.Done()
	fmt.Println("\nStopped watching positions.")
	return nil
}

func cmdWatchKlines(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "watch-klines <symbol> <interval>"); err != nil {
		return err
	}
	symbol := args[0]
	interval := exchanges.Interval(args[1])

	fmt.Printf("Watching %s klines for %s (Ctrl+C to stop)...\n", interval, symbol)

	err := adp.WatchKlines(ctx, symbol, interval, func(k *exchanges.Kline) {
		if jsonOut {
			outputJSON(k)
			return
		}
		ts := time.UnixMilli(k.Timestamp).Format("15:04:05")
		change := k.Close.Sub(k.Open)
		dir := colorGreen("▲")
		if change.IsNegative() {
			dir = colorRed("▼")
		} else if change.IsZero() {
			dir = colorDim("─")
		}
		fmt.Printf("[%s] %s O: %s  H: %s  L: %s  C: %s %s  Vol: %s\n",
			ts,
			symbol,
			decStr(k.Open),
			decStr(k.High),
			decStr(k.Low),
			decStr(k.Close),
			dir,
			decStr(k.Volume),
		)
	})
	if err != nil {
		return fmt.Errorf("watch klines: %w", err)
	}

	<-ctx.Done()
	fmt.Println("\nStopped watching klines.")
	return nil
}

// trimOrderBook trims the order book to the requested depth.
func trimOrderBook(ob *exchanges.OrderBook, depth int) *exchanges.OrderBook {
	trimmed := &exchanges.OrderBook{
		Symbol:    ob.Symbol,
		Timestamp: ob.Timestamp,
	}
	if len(ob.Asks) > depth {
		trimmed.Asks = ob.Asks[:depth]
	} else {
		trimmed.Asks = ob.Asks
	}
	if len(ob.Bids) > depth {
		trimmed.Bids = ob.Bids[:depth]
	} else {
		trimmed.Bids = ob.Bids
	}
	return trimmed
}

// printLiveOrderBook prints the order book in a compact live format.
func printLiveOrderBook(ob *exchanges.OrderBook, depth int) {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")
	fmt.Printf("=== %s Order Book ===\n\n", ob.Symbol)

	askEnd := depth
	if askEnd > len(ob.Asks) {
		askEnd = len(ob.Asks)
	}

	// Find max quantity for proportional bars
	maxQty := findMaxQty(ob.Asks, ob.Bids, depth)

	fmt.Println(colorRed("  ASKS"))
	for i := askEnd - 1; i >= 0; i-- {
		bar := quantityBar(ob.Asks[i].Quantity, maxQty, 20)
		fmt.Printf("  %s  %s  %s\n",
			colorRed(fmt.Sprintf("%-14s", decStr(ob.Asks[i].Price))),
			fmt.Sprintf("%-14s", decStr(ob.Asks[i].Quantity)),
			colorRed(bar),
		)
	}

	// Spread indicator
	if askEnd > 0 && len(ob.Bids) > 0 {
		spread := ob.Asks[0].Price.Sub(ob.Bids[0].Price)
		fmt.Printf("  ──── spread: %s ────\n", decStr(spread))
	} else {
		fmt.Println("  ────────────────────────────────")
	}

	bidEnd := depth
	if bidEnd > len(ob.Bids) {
		bidEnd = len(ob.Bids)
	}

	fmt.Println(colorGreen("  BIDS"))
	for i := 0; i < bidEnd; i++ {
		bar := quantityBar(ob.Bids[i].Quantity, maxQty, 20)
		fmt.Printf("  %s  %s  %s\n",
			colorGreen(fmt.Sprintf("%-14s", decStr(ob.Bids[i].Price))),
			fmt.Sprintf("%-14s", decStr(ob.Bids[i].Quantity)),
			colorGreen(bar),
		)
	}

	fmt.Printf("\n  Last update: %s\n", time.Now().Format("15:04:05.000"))
}

// findMaxQty finds the maximum quantity across asks and bids up to given depth.
func findMaxQty(asks, bids []exchanges.Level, depth int) decimal.Decimal {
	maxQty := decimal.Zero
	for i := 0; i < depth && i < len(asks); i++ {
		if asks[i].Quantity.GreaterThan(maxQty) {
			maxQty = asks[i].Quantity
		}
	}
	for i := 0; i < depth && i < len(bids); i++ {
		if bids[i].Quantity.GreaterThan(maxQty) {
			maxQty = bids[i].Quantity
		}
	}
	return maxQty
}

// quantityBar creates a proportional bar string.
func quantityBar(qty, maxQty decimal.Decimal, maxLen int) string {
	if maxQty.IsZero() || qty.IsZero() {
		return ""
	}
	ratio := qty.Div(maxQty)
	barLen := ratio.Mul(decimal.NewFromInt(int64(maxLen))).IntPart()
	if barLen < 1 {
		barLen = 1
	}
	if barLen > int64(maxLen) {
		barLen = int64(maxLen)
	}
	return strings.Repeat("█", int(barLen))
}
