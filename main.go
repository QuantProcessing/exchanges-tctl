package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	exchanges "github.com/QuantProcessing/exchanges"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	// Load .env file (optional)
	_ = godotenv.Load()

	// Global flags
	exchange := flag.String("e", "", "exchange name (e.g. BINANCE, OKX). Auto-detected if only one is configured.")
	market := flag.String("m", "perp", "market type: perp | spot")
	jsonOut := flag.Bool("json", false, "output as JSON (for AI agents)")
	useWS := flag.Bool("ws", false, "use WebSocket for order operations (default: REST)")
	showVersion := flag.Bool("version", false, "show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `tctl — trader exchange control tool

Usage:
  tctl [flags] <command> [args...]

Global Flags:
  -e string     exchange name (auto-detected if only one configured)
  -m string     market type: perp | spot (default "perp")
  -json         output as JSON
  -ws           use WebSocket for order operations
  -version      show version

Market Data:
  ticker <symbol>                          get ticker price
  orderbook <symbol> [depth]               get order book
  trades <symbol> [limit]                  recent market trades
  klines <symbol> <interval> [limit]       candlestick data
  details <symbol>                         symbol details
  fee <symbol>                             fee rate
  funding <symbol>                         funding rate (perp only)

Trading:
  buy <symbol> <qty> [--price P] [flags]   place buy order
  sell <symbol> <qty> [--price P] [flags]  place sell order
  modify <orderID> <symbol> [flags]        modify order (perp only)
  cancel <orderID> <symbol>                cancel order
  cancel-all <symbol>                      cancel all orders
  order <orderID> <symbol>                 fetch order details

  Order flags: --price P --tif GTC|IOC|FOK|PO --post-only --reduce-only --client-id ID

Account:
  positions                                list positions (perp only)
  orders [symbol]                          list open orders
  balance                                  show balance
  account                                  full account info
  leverage <symbol> <value>                set leverage (perp only)
  spot-balances                            spot asset balances (spot only)
  transfer <asset> <amount> [flags]        asset transfer (spot only)

Streaming (requires -ws):
  watch-ticker <symbol>                    live ticker updates
  watch-ob <symbol> [depth]                live order book
  watch-orders                             live order updates
  watch-trades <symbol>                    live trade stream

Aliases:
  t=ticker  ob=orderbook  kl=klines  p=positions  o=orders
  b=balance  acc=account  sb=spot-balances  lev=leverage
  wt=watch-ticker  wob=watch-ob  wo=watch-orders  wtr=watch-trades

Examples:
  tctl ticker BTC
  tctl -m spot -e BINANCE spot-balances
  tctl buy BTC 0.001 --price 50000
  tctl buy ETH 0.5 --price 3000 --tif IOC
  tctl sell BTC 0.01 --price 95000 --post-only
  tctl order order-123 BTC
  tctl -ws watch-ticker ETH
  tctl positions --json
`)
	}

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// Initialize logger (quiet by default for CLI tool)
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	args := flag.Args()
	if len(args) == 0 {
		// Interactive mode
		runInteractive(*exchange, *market, *jsonOut, *useWS, logger)
		return
	}

	// One-shot mode
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Auto-detect exchange
	exchName := resolveExchange(*exchange)
	if exchName == "" {
		fatal(*jsonOut, "no exchange specified and none auto-detected. Use -e flag or configure credentials in .env")
	}

	// Determine market type
	marketType := exchanges.MarketTypePerp
	if *market == "spot" {
		marketType = exchanges.MarketTypeSpot
	}

	// Create adapter
	var adp exchanges.Exchange
	var err error
	if *useWS {
		adp, err = createAdapter(ctx, exchName, marketType, logger)
	} else {
		adp, err = createRESTAdapter(ctx, exchName, marketType, logger)
	}
	if err != nil {
		fatal(*jsonOut, "failed to create %s adapter: %v", exchName, err)
	}

	// Dispatch command
	cmd := strings.ToLower(args[0])
	cmdArgs := args[1:]

	if err := dispatch(ctx, adp, exchName, cmd, cmdArgs, *jsonOut); err != nil {
		fatal(*jsonOut, "%v", err)
	}
}

// fatal outputs an error and exits
func fatal(jsonOut bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if jsonOut {
		b, _ := json.Marshal(map[string]string{"error": msg})
		fmt.Fprintln(os.Stdout, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	os.Exit(1)
}
