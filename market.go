package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	exchanges "github.com/QuantProcessing/exchanges"
)

func cmdTicker(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "ticker <symbol>"); err != nil {
		return err
	}
	symbol := args[0]
	ticker, err := adp.FetchTicker(ctx, symbol)
	if err != nil {
		return fmt.Errorf("fetch ticker: %w", err)
	}

	if jsonOut {
		outputJSON(ticker)
		return nil
	}

	outputTable(
		[]string{"Symbol", "Bid", "Ask", "Last", "Volume"},
		[][]string{{
			ticker.Symbol,
			colorGreen(decStr(ticker.Bid)),
			colorRed(decStr(ticker.Ask)),
			decStr(ticker.LastPrice),
			decStr(ticker.Volume24h),
		}},
	)
	return nil
}

func cmdOrderBook(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "orderbook <symbol> [depth]"); err != nil {
		return err
	}
	symbol := args[0]
	depth := 10
	if len(args) > 1 {
		if d, err := parseInt(args[1]); err == nil {
			depth = d
		}
	}

	ob, err := adp.FetchOrderBook(ctx, symbol, depth)
	if err != nil {
		return fmt.Errorf("fetch orderbook: %w", err)
	}

	if jsonOut {
		outputJSON(ob)
		return nil
	}

	fmt.Printf("=== %s Order Book ===\n", symbol)
	fmt.Println("\n" + colorRed("Asks (sell):"))
	outputTable([]string{"Price", "Quantity"}, func() [][]string {
		rows := make([][]string, 0, len(ob.Asks))
		// Show asks in reverse (best ask at bottom)
		for i := len(ob.Asks) - 1; i >= 0; i-- {
			rows = append(rows, []string{colorRed(decStr(ob.Asks[i].Price)), decStr(ob.Asks[i].Quantity)})
		}
		return rows
	}())
	fmt.Println("\n" + colorGreen("Bids (buy):"))
	outputTable([]string{"Price", "Quantity"}, func() [][]string {
		rows := make([][]string, 0, len(ob.Bids))
		for _, b := range ob.Bids {
			rows = append(rows, []string{colorGreen(decStr(b.Price)), decStr(b.Quantity)})
		}
		return rows
	}())
	return nil
}

func cmdSymbolDetails(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "details <symbol>"); err != nil {
		return err
	}
	details, err := adp.FetchSymbolDetails(ctx, args[0])
	if err != nil {
		return fmt.Errorf("fetch symbol details: %w", err)
	}

	if jsonOut {
		outputJSON(details)
		return nil
	}

	outputTable(
		[]string{"Field", "Value"},
		[][]string{
			{"Symbol", details.Symbol},
			{"Price Precision", fmt.Sprintf("%d", details.PricePrecision)},
			{"Qty Precision", fmt.Sprintf("%d", details.QuantityPrecision)},
			{"Min Quantity", decStr(details.MinQuantity)},
			{"Min Notional", decStr(details.MinNotional)},
		},
	)
	return nil
}

func cmdFeeRate(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "fee <symbol>"); err != nil {
		return err
	}
	fee, err := adp.FetchFeeRate(ctx, args[0])
	if err != nil {
		return fmt.Errorf("fetch fee rate: %w", err)
	}

	if jsonOut {
		outputJSON(fee)
		return nil
	}

	outputTable(
		[]string{"Maker", "Taker"},
		[][]string{{
			decStr(fee.Maker),
			decStr(fee.Taker),
		}},
	)
	return nil
}

func cmdFundingRate(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "funding <symbol>"); err != nil {
		return err
	}
	perpAdp, ok := adp.(exchanges.PerpExchange)
	if !ok {
		return fmt.Errorf("funding rate is only available for perpetual contracts (current: %s mode)", adp.GetMarketType())
	}
	rate, err := perpAdp.FetchFundingRate(ctx, args[0])
	if err != nil {
		return fmt.Errorf("fetch funding rate: %w", err)
	}

	if jsonOut {
		outputJSON(rate)
		return nil
	}

	nextFunding := time.Unix(rate.NextFundingTime/1000, 0).Format("2006-01-02 15:04:05")
	outputTable(
		[]string{"Symbol", "Rate", "Next Funding"},
		[][]string{{
			rate.Symbol,
			decStr(rate.FundingRate),
			nextFunding,
		}},
	)
	return nil
}

func cmdTrades(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 1, "trades <symbol> [limit]"); err != nil {
		return err
	}
	symbol := args[0]
	limit := 20
	if len(args) > 1 {
		if l, err := parseInt(args[1]); err == nil {
			limit = l
		}
	}

	trades, err := adp.FetchTrades(ctx, symbol, limit)
	if err != nil {
		return fmt.Errorf("fetch trades: %w", err)
	}

	if jsonOut {
		outputJSON(trades)
		return nil
	}

	if len(trades) == 0 {
		outputInfo("No recent trades for %s", symbol)
		return nil
	}

	rows := make([][]string, 0, len(trades))
	for _, t := range trades {
		ts := time.UnixMilli(t.Timestamp).Format("15:04:05")
		rows = append(rows, []string{
			ts,
			colorSide(strings.ToUpper(string(t.Side))),
			decStr(t.Price),
			decStr(t.Quantity),
		})
	}
	outputTable([]string{"Time", "Side", "Price", "Qty"}, rows)
	return nil
}

func cmdKlines(ctx context.Context, adp exchanges.Exchange, args []string, jsonOut bool) error {
	if err := requireArgs(args, 2, "klines <symbol> <interval> [limit]"); err != nil {
		return err
	}
	symbol := args[0]
	interval := exchanges.Interval(args[1])
	limit := 20
	if len(args) > 2 {
		if l, err := parseInt(args[2]); err == nil {
			limit = l
		}
	}

	opts := &exchanges.KlineOpts{Limit: limit}
	klines, err := adp.FetchKlines(ctx, symbol, interval, opts)
	if err != nil {
		return fmt.Errorf("fetch klines: %w", err)
	}

	if jsonOut {
		outputJSON(klines)
		return nil
	}

	if len(klines) == 0 {
		outputInfo("No klines data for %s %s", symbol, interval)
		return nil
	}

	rows := make([][]string, 0, len(klines))
	for _, k := range klines {
		ts := time.UnixMilli(k.Timestamp).Format("01-02 15:04")
		// Color the close based on open-close comparison
		closeStr := decStr(k.Close)
		if k.Close.GreaterThan(k.Open) {
			closeStr = colorGreen(closeStr)
		} else if k.Close.LessThan(k.Open) {
			closeStr = colorRed(closeStr)
		}
		rows = append(rows, []string{
			ts,
			decStr(k.Open),
			decStr(k.High),
			decStr(k.Low),
			closeStr,
			decStr(k.Volume),
		})
	}
	outputTable([]string{"Time", "Open", "High", "Low", "Close", "Volume"}, rows)
	return nil
}

func cmdAllFundingRates(ctx context.Context, adp exchanges.Exchange, jsonOut bool) error {
	perpAdp, ok := adp.(exchanges.PerpExchange)
	if !ok {
		return fmt.Errorf("funding-all is only available for perpetual contracts (current: %s mode)", adp.GetMarketType())
	}

	rates, err := perpAdp.FetchAllFundingRates(ctx)
	if err != nil {
		return fmt.Errorf("fetch all funding rates: %w", err)
	}

	if jsonOut {
		outputJSON(rates)
		return nil
	}

	rows := make([][]string, 0, len(rates))
	for _, r := range rates {
		rows = append(rows, []string{
			r.Symbol,
			colorPnL(r.FundingRate),
			time.Unix(r.NextFundingTime/1000, 0).Format("15:04:05"),
		})
	}
	outputTable([]string{"Symbol", "Rate", "Next"}, rows)
	return nil
}
