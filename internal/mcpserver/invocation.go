package mcpserver

import (
	"fmt"
	"strconv"
	"strings"
)

func BuildCLIInvocation(req ToolCallRequest) (CLIInvocation, error) {
	tool, ok := ToolByName(req.ToolName)
	if !ok {
		return CLIInvocation{}, fmt.Errorf("unknown tool: %s", req.ToolName)
	}

	args := make([]string, 0, 16)
	if exch := strings.TrimSpace(req.Exchange); exch != "" {
		args = append(args, "-e", exch)
	}
	if market := strings.TrimSpace(req.Market); market != "" {
		args = append(args, "-m", market)
	}
	args = append(args, "-json")

	commandArgs, err := buildCommandArgs(tool.Command, req.Arguments)
	if err != nil {
		return CLIInvocation{}, err
	}
	args = append(args, commandArgs...)

	return CLIInvocation{
		Command: currentTCTLBinary(),
		Args:    args,
	}, nil
}

func buildCommandArgs(command string, args map[string]any) ([]string, error) {
	switch command {
	case "ticker":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		return []string{"ticker", symbol}, nil
	case "buy", "sell":
		return buildPlaceOrderArgs(command, args)
	case "fund-arb":
		return buildFundArbArgs(args)
	case "orderbook":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		out := []string{"orderbook", symbol}
		if depth, ok := optionalStringArg(args, "depth"); ok {
			out = append(out, depth)
		}
		return out, nil
	case "details", "fee", "funding":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		return []string{command, symbol}, nil
	case "trades":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		out := []string{"trades", symbol}
		if limit, ok := optionalStringArg(args, "limit"); ok {
			out = append(out, limit)
		}
		return out, nil
	case "klines":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		interval, err := requiredStringArg(args, "interval")
		if err != nil {
			return nil, err
		}
		out := []string{"klines", symbol, interval}
		if limit, ok := optionalStringArg(args, "limit"); ok {
			out = append(out, limit)
		}
		return out, nil
	case "order", "cancel":
		orderID, err := requiredStringArg(args, "order_id")
		if err != nil {
			return nil, err
		}
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		return []string{command, orderID, symbol}, nil
	case "orders":
		out := []string{"orders"}
		if symbol, ok := optionalStringArg(args, "symbol"); ok {
			out = append(out, symbol)
		}
		return out, nil
	case "balance", "account", "positions", "funding-all", "spot-balances":
		return []string{command}, nil
	case "cancel-all":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		return []string{"cancel-all", symbol}, nil
	case "modify":
		orderID, err := requiredStringArg(args, "order_id")
		if err != nil {
			return nil, err
		}
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		out := []string{"modify", orderID, symbol}
		if price, ok := optionalStringArg(args, "price"); ok {
			out = append(out, "--price", price)
		}
		if qty, ok := optionalStringArg(args, "quantity"); ok {
			out = append(out, "--qty", qty)
		}
		return out, nil
	case "leverage":
		symbol, err := requiredStringArg(args, "symbol")
		if err != nil {
			return nil, err
		}
		lev, err := requiredStringArg(args, "leverage")
		if err != nil {
			return nil, err
		}
		return []string{"leverage", symbol, lev}, nil
	case "transfer":
		asset, err := requiredStringArg(args, "asset")
		if err != nil {
			return nil, err
		}
		amount, err := requiredStringArg(args, "amount")
		if err != nil {
			return nil, err
		}
		out := []string{"transfer", asset, amount}
		if from, ok := optionalStringArg(args, "from"); ok {
			out = append(out, "--from", from)
		}
		if to, ok := optionalStringArg(args, "to"); ok {
			out = append(out, "--to", to)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported command mapping: %s", command)
	}
}

func buildPlaceOrderArgs(command string, args map[string]any) ([]string, error) {
	symbol, err := requiredStringArg(args, "symbol")
	if err != nil {
		return nil, err
	}
	qty, err := requiredStringArg(args, "quantity")
	if err != nil {
		return nil, err
	}
	out := []string{command, symbol, qty}
	if price, ok := optionalStringArg(args, "price"); ok {
		out = append(out, "--price", price)
	}
	if tif, ok := optionalStringArg(args, "tif"); ok {
		out = append(out, "--tif", tif)
	}
	if flagIsTrue(args, "post_only") {
		out = append(out, "--post-only")
	}
	if flagIsTrue(args, "reduce_only") {
		out = append(out, "--reduce-only")
	}
	if clientID, ok := optionalStringArg(args, "client_id"); ok {
		out = append(out, "--client-id", clientID)
	}
	return out, nil
}

func buildFundArbArgs(args map[string]any) ([]string, error) {
	symbol, err := requiredStringArg(args, "symbol")
	if err != nil {
		return nil, err
	}
	qty, err := requiredStringArg(args, "quantity")
	if err != nil {
		return nil, err
	}
	out := []string{"fund-arb", symbol, qty}
	if spotExchange, ok := optionalStringArg(args, "spot_exchange"); ok {
		out = append(out, "--spot-exchange", spotExchange)
	}
	if perpExchange, ok := optionalStringArg(args, "perp_exchange"); ok {
		out = append(out, "--perp-exchange", perpExchange)
	}
	if leverage, ok := optionalStringArg(args, "leverage"); ok {
		out = append(out, "--leverage", leverage)
	}
	if flagIsTrue(args, "close") {
		out = append(out, "--close")
	}
	if spotPrice, ok := optionalStringArg(args, "spot_price"); ok {
		out = append(out, "--spot-price", spotPrice)
	}
	if perpPrice, ok := optionalStringArg(args, "perp_price"); ok {
		out = append(out, "--perp-price", perpPrice)
	}
	return out, nil
}

func requiredStringArg(args map[string]any, key string) (string, error) {
	val, ok := optionalStringArg(args, key)
	if !ok {
		return "", fmt.Errorf("missing required argument: %s", key)
	}
	return val, nil
}

func optionalStringArg(args map[string]any, key string) (string, bool) {
	if args == nil {
		return "", false
	}
	value, ok := args[key]
	if !ok || value == nil {
		return "", false
	}
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		return v, true
	case int:
		return strconv.Itoa(v), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10), true
		}
		return strconv.FormatFloat(v, 'f', -1, 64), true
	default:
		return fmt.Sprint(v), true
	}
}

func flagIsTrue(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	value, ok := args[key]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		}
	}
	return false
}
