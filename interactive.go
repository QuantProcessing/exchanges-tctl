package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	exchanges "github.com/QuantProcessing/exchanges"
	"go.uber.org/zap"
)

// interactiveState holds mutable state for the interactive session.
type interactiveState struct {
	exchName   string
	market     string
	marketType exchanges.MarketType
	useWS      bool
	jsonOut    bool
	adp        exchanges.Exchange
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.SugaredLogger
}

func (s *interactiveState) prompt() string {
	mode := "rest"
	if s.useWS {
		mode = "ws"
	}
	return fmt.Sprintf("%s/%s(%s)> ", colorBold(s.exchName), s.market, mode)
}

func (s *interactiveState) reconnect() error {
	// Close existing adapter if any
	if s.adp != nil {
		s.adp.Close()
	}

	var err error
	if s.useWS {
		s.adp, err = createAdapter(s.ctx, s.exchName, s.marketType, s.logger)
	} else {
		s.adp, err = createRESTAdapter(s.ctx, s.exchName, s.marketType, s.logger)
	}
	return err
}

func runInteractive(exchangeFlag, market string, jsonOut, useWS bool, logger *zap.SugaredLogger) {
	exchName := resolveExchange(exchangeFlag)
	if exchName == "" {
		// Show available exchanges and prompt
		configured := configuredExchanges()
		if len(configured) == 0 {
			fatal(jsonOut, "no exchanges configured. Set credentials in .env file.")
		}
		fmt.Printf("Available exchanges: %s\n", colorCyan(strings.Join(configured, ", ")))
		fmt.Print("Select exchange: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			exchName = strings.ToUpper(strings.TrimSpace(scanner.Text()))
		}
		if exchName == "" {
			return
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	marketType := exchanges.MarketTypePerp
	if market == "spot" {
		marketType = exchanges.MarketTypeSpot
	}

	state := &interactiveState{
		exchName:   exchName,
		market:     market,
		marketType: marketType,
		useWS:      useWS,
		jsonOut:    jsonOut,
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}

	if err := state.reconnect(); err != nil {
		fatal(jsonOut, "failed to create %s adapter: %v", exchName, err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Connected to %s (%s). Type %s for commands, %s to quit.\n",
		colorBold(exchName), market, colorCyan("help"), colorCyan("exit"))

	for {
		fmt.Print(state.prompt())
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit", "q":
			fmt.Println("Bye!")
			return
		case "help", "h", "?":
			printInteractiveHelp()
			continue
		case "status":
			printStatus(state)
			continue
		case "use", "switch":
			if len(parts) < 2 {
				outputError("Usage: use <exchange>")
				continue
			}
			newExch := strings.ToUpper(parts[1])
			state.exchName = newExch
			if err := state.reconnect(); err != nil {
				outputError("Failed to connect to %s: %v", newExch, err)
				continue
			}
			outputSuccess("Switched to %s", newExch)
			continue
		case "market":
			if len(parts) < 2 {
				outputError("Usage: market perp|spot")
				continue
			}
			newMarket := strings.ToLower(parts[1])
			switch newMarket {
			case "perp":
				state.market = "perp"
				state.marketType = exchanges.MarketTypePerp
			case "spot":
				state.market = "spot"
				state.marketType = exchanges.MarketTypeSpot
			default:
				outputError("Invalid market type. Use: perp | spot")
				continue
			}
			if err := state.reconnect(); err != nil {
				outputError("Failed to switch to %s: %v", newMarket, err)
				continue
			}
			outputSuccess("Switched to %s market", newMarket)
			continue
		case "mode":
			if len(parts) < 2 {
				outputError("Usage: mode rest|ws")
				continue
			}
			switch strings.ToLower(parts[1]) {
			case "rest":
				state.useWS = false
			case "ws":
				state.useWS = true
			default:
				outputError("Invalid mode. Use: rest | ws")
				continue
			}
			if err := state.reconnect(); err != nil {
				outputError("Failed to switch mode: %v", err)
				continue
			}
			outputSuccess("Switched to %s mode", parts[1])
			continue
		case "json":
			state.jsonOut = !state.jsonOut
			if state.jsonOut {
				outputSuccess("JSON output enabled")
			} else {
				outputSuccess("JSON output disabled")
			}
			continue
		}

		// fund-arb is special: creates its own spot+perp adapters
		if cmd == "fund-arb" || cmd == "fa" {
			if err := cmdFundArb(state.ctx, state.exchName, parts[1:], state.jsonOut, state.logger); err != nil {
				outputError("%v", err)
			}
			continue
		}

		if err := dispatch(state.ctx, state.adp, state.exchName, cmd, parts[1:], state.jsonOut); err != nil {
			outputError("%v", err)
		}
	}
}

func printStatus(s *interactiveState) {
	mode := "REST"
	if s.useWS {
		mode = "WebSocket"
	}
	fmt.Printf("  Exchange:  %s\n", colorBold(s.exchName))
	fmt.Printf("  Market:    %s\n", s.market)
	fmt.Printf("  Mode:      %s\n", mode)
	fmt.Printf("  JSON:      %v\n", s.jsonOut)

	configured := configuredExchanges()
	fmt.Printf("  Available: %s\n", colorDim(strings.Join(configured, ", ")))
}

func printInteractiveHelp() {
	fmt.Println(colorBold("Market Data:"))
	fmt.Println("  ticker <symbol>              (t)    get ticker")
	fmt.Println("  orderbook <symbol> [depth]    (ob)   get order book")
	fmt.Println("  trades <symbol> [limit]              recent trades")
	fmt.Println("  klines <symbol> <interval>    (kl)   candlestick data")
	fmt.Println("  details <symbol>                     symbol details")
	fmt.Println("  fee <symbol>                         fee rate")
	fmt.Println("  funding <symbol>                     funding rate (perp)")
	fmt.Println()
	fmt.Println(colorBold("Trading:"))
	fmt.Println("  buy <symbol> <qty> [--price P] [flags]       place buy order")
	fmt.Println("  sell <symbol> <qty> [--price P] [flags]      place sell order")
	fmt.Println("  modify <orderID> <symbol> [flags]            modify order (perp)")
	fmt.Println("  cancel <orderID> <symbol>                    cancel order")
	fmt.Println("  cancel-all <symbol>                          cancel all orders")
	fmt.Println("  order <orderID> <symbol>                     fetch order details")
	fmt.Println("  Flags: --tif GTC|IOC|FOK|PO --post-only --reduce-only --client-id ID")
	fmt.Println()
	fmt.Println(colorBold("Account:"))
	fmt.Println("  positions                     (p)    list positions (perp)")
	fmt.Println("  orders [symbol]               (o)    list open orders")
	fmt.Println("  balance                       (b)    show balance")
	fmt.Println("  account                       (acc)  full account info")
	fmt.Println("  leverage <symbol> <value>     (lev)  set leverage (perp)")
	fmt.Println("  spot-balances                 (sb)   spot balances (spot)")
	fmt.Println("  transfer <asset> <amount> [flags]    asset transfer (spot)")
	fmt.Println()
	fmt.Println(colorBold("Streaming (WSS):"))
	fmt.Println("  watch-ticker <symbol>         (wt)   live ticker")
	fmt.Println("  watch-ob <symbol> [depth]     (wob)  live order book")
	fmt.Println("  watch-orders                  (wo)   live order updates")
	fmt.Println("  watch-trades <symbol>         (wtr)  live trade stream")
	fmt.Println()
	fmt.Println(colorBold("Session:"))
	fmt.Println("  use <exchange>                       switch exchange")
	fmt.Println("  market perp|spot                     switch market type")
	fmt.Println("  mode rest|ws                         switch transport mode")
	fmt.Println("  status                               show session info")
	fmt.Println("  json                                 toggle JSON output")
	fmt.Println("  help                          (h,?)  show this help")
	fmt.Println("  exit                          (q)    quit")
}
