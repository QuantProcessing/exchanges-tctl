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
	"github.com/QuantProcessing/exchanges-tctl/internal/mcpserver"
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
	useWS := flag.Bool("ws", false, "compatibility flag for watch/WebSocket-centric workflows")
	showVersion := flag.Bool("version", false, "show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `tctl — trader exchange control tool

Usage:
  tctl [flags] <command> [args...]

Global Flags:
  -e string     exchange name (auto-detected if only one configured)
  -m string     market type: perp | spot (default "perp")
  -json         output as JSON
  -ws           compatibility flag for watch/WebSocket-centric workflows
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

Arbitrage:
  fund-arb <symbol> <qty> [flags]          funding rate arb (buy spot + short perp)
  Fund-arb flags: --leverage N --close --spot-price P --perp-price P
                  [--spot-exchange EX --perp-exchange EX]

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

MCP:
  mcp serve                                start the stdio MCP server

Aliases:
  t=ticker  ob=orderbook  kl=klines  p=positions  o=orders
  b=balance  acc=account  sb=spot-balances  lev=leverage
  wt=watch-ticker  wob=watch-ob  wo=watch-orders  wtr=watch-trades
  fa=fund-arb

Examples:
  tctl ticker BTC
  tctl -m spot -e BINANCE spot-balances
  tctl buy BTC 0.001 --price 50000
  tctl buy ETH 0.5 --price 3000 --tif IOC
  tctl sell BTC 0.01 --price 95000 --post-only
  tctl order order-123 BTC
  tctl -ws watch-ticker ETH
  tctl positions --json
  tctl fund-arb BTC 0.01 --leverage 5
  tctl fund-arb BTC 0.01 --close
  tctl fund-arb BTC 0.01 --spot-exchange BINANCE --perp-exchange OKX
  tctl mcp serve
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

	if isMCPServeCommand(args) {
		runMCPServer(ctx)
		return
	}

	cmd := strings.ToLower(args[0])
	cmdArgs := args[1:]

	// fund-arb is special: it creates its own spot+perp adapters internally
	// and can resolve exchanges per leg, so handle it before global exchange setup.
	if cmd == "fund-arb" || cmd == "fa" {
		if err := cmdFundArb(ctx, *exchange, cmdArgs, *jsonOut, logger); err != nil {
			fatal(*jsonOut, "%v", err)
		}
		return
	}

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

	if err := dispatch(ctx, adp, exchName, cmd, cmdArgs, *jsonOut); err != nil {
		fatal(*jsonOut, "%v", err)
	}
}

func isMCPServeCommand(args []string) bool {
	if len(args) < 2 {
		return false
	}
	return strings.EqualFold(args[0], "mcp") && strings.EqualFold(args[1], "serve")
}

func runMCPServer(ctx context.Context) {
	server := mcpserver.NewServer(mcpserver.ServerOptions{
		Name:    "tctl",
		Version: version,
	})
	if err := server.ServeStdio(ctx, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
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
