package main

import (
	"context"
	"strings"
	"testing"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

func TestCmdWatchTickerRequiresArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdWatchTicker(ctx, mock, nil, false)
	if err == nil {
		t.Error("watch-ticker without args should fail")
	}
}

func TestCmdWatchTickerDispatch(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		err := cmdWatchTicker(cancelCtx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "WatchTicker" {
		t.Errorf("expected WatchTicker, got %s", mock.lastMethod)
	}
}

func TestCmdWatchOrderBookRequiresArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdWatchOrderBook(ctx, mock, nil, false)
	if err == nil {
		t.Error("watch-ob without args should fail")
	}
}

func TestCmdWatchOrderBookDispatch(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		cmdWatchOrderBook(cancelCtx, mock, []string{"BTC"}, false)
	})

	if mock.lastMethod != "WatchOrderBook" {
		t.Errorf("expected WatchOrderBook, got %s", mock.lastMethod)
	}
}

func TestCmdWatchOrdersDispatch(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		cmdWatchOrders(cancelCtx, mock, nil, false)
	})

	if mock.lastMethod != "WatchOrders" {
		t.Errorf("expected WatchOrders, got %s", mock.lastMethod)
	}
}

func TestCmdWatchTradesRequiresArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdWatchTrades(ctx, mock, nil, false)
	if err == nil {
		t.Error("watch-trades without args should fail")
	}
}

func TestCmdWatchTradesDispatch(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		cmdWatchTrades(cancelCtx, mock, []string{"BTC"}, false)
	})

	if mock.lastMethod != "WatchTrades" {
		t.Errorf("expected WatchTrades, got %s", mock.lastMethod)
	}
}

func TestTrimOrderBook(t *testing.T) {
	ob := &exchanges.OrderBook{
		Symbol: "BTC",
		Asks:   make([]exchanges.Level, 20),
		Bids:   make([]exchanges.Level, 20),
	}

	trimmed := trimOrderBook(ob, 5)
	if len(trimmed.Asks) != 5 {
		t.Errorf("trimmed asks = %d, want 5", len(trimmed.Asks))
	}
	if len(trimmed.Bids) != 5 {
		t.Errorf("trimmed bids = %d, want 5", len(trimmed.Bids))
	}
}

func TestTrimOrderBookSmall(t *testing.T) {
	ob := &exchanges.OrderBook{
		Symbol: "BTC",
		Asks:   make([]exchanges.Level, 3),
		Bids:   make([]exchanges.Level, 2),
	}

	trimmed := trimOrderBook(ob, 10)
	if len(trimmed.Asks) != 3 {
		t.Errorf("trimmed asks = %d, want 3", len(trimmed.Asks))
	}
	if len(trimmed.Bids) != 2 {
		t.Errorf("trimmed bids = %d, want 2", len(trimmed.Bids))
	}
}

// ============================================================================
// WSS streaming alias tests
// ============================================================================

func TestStreamingAliases(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	tests := []struct {
		alias      string
		args       []string
		wantMethod string
	}{
		{"wt", []string{"BTC"}, "WatchTicker"},
		{"wob", []string{"BTC"}, "WatchOrderBook"},
		{"wo", nil, "WatchOrders"},
		{"wtr", []string{"BTC"}, "WatchTrades"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			mock.lastMethod = ""
			captureOutput(func() {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				_ = dispatch(cancelCtx, mock, "MOCK", tt.alias, tt.args, true)
			})
			if mock.lastMethod != tt.wantMethod {
				t.Errorf("alias %q called %q, want %q", tt.alias, mock.lastMethod, tt.wantMethod)
			}
		})
	}
}

func TestStreamingArgsValidation(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	cmds := []string{"watch-ticker", "watch-ob", "watch-trades", "wt", "wob", "wtr"}
	for _, cmd := range cmds {
		t.Run(cmd+"_no_args", func(t *testing.T) {
			err := dispatch(ctx, mock, "MOCK", cmd, nil, false)
			if err == nil {
				t.Errorf("dispatch(%q, nil) should fail due to missing args", cmd)
			}
			if !strings.Contains(err.Error(), "not enough arguments") {
				t.Errorf("error should mention args, got: %v", err)
			}
		})
	}
}

// ============================================================================
// quantityBar and findMaxQty tests
// ============================================================================

func TestQuantityBar(t *testing.T) {
	tests := []struct {
		name   string
		qty    string
		maxQty string
		maxLen int
		want   int
	}{
		{"full bar", "100", "100", 20, 20},
		{"half bar", "50", "100", 20, 10},
		{"quarter bar", "25", "100", 20, 5},
		{"minimum 1", "1", "1000", 20, 1},
		{"zero qty", "0", "100", 20, 0},
		{"zero max", "50", "0", 20, 0},
		{"small ratio", "0.001", "100", 20, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qty := decimal.RequireFromString(tt.qty)
			maxQty := decimal.RequireFromString(tt.maxQty)
			bar := quantityBar(qty, maxQty, tt.maxLen)
			got := len([]rune(bar))
			if tt.want == 0 {
				if bar != "" {
					t.Errorf("quantityBar() = %q (len %d), want empty", bar, got)
				}
			} else if got != tt.want {
				t.Errorf("quantityBar() len = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFindMaxQty(t *testing.T) {
	asks := []exchanges.Level{
		{Quantity: decimal.RequireFromString("10")},
		{Quantity: decimal.RequireFromString("50")},
	}
	bids := []exchanges.Level{
		{Quantity: decimal.RequireFromString("30")},
		{Quantity: decimal.RequireFromString("20")},
	}

	maxQty := findMaxQty(asks, bids, 5)
	if !maxQty.Equal(decimal.RequireFromString("50")) {
		t.Errorf("findMaxQty() = %s, want 50", maxQty.String())
	}
}

func TestFindMaxQtyEmpty(t *testing.T) {
	maxQty := findMaxQty(nil, nil, 5)
	if !maxQty.IsZero() {
		t.Errorf("findMaxQty(nil, nil) = %s, want 0", maxQty.String())
	}
}

func TestFindMaxQtyDepthLimit(t *testing.T) {
	levels := []exchanges.Level{
		{Quantity: decimal.RequireFromString("10")},
		{Quantity: decimal.RequireFromString("100")},
	}

	maxQty := findMaxQty(levels, nil, 1)
	if !maxQty.Equal(decimal.RequireFromString("10")) {
		t.Errorf("findMaxQty(depth=1) = %s, want 10", maxQty.String())
	}
}
