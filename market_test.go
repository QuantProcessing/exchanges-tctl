package main

import (
	"context"
	"strings"
	"testing"
)

func TestCmdTicker(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTicker(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "FetchTicker" {
		t.Errorf("expected FetchTicker, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "95000") {
		t.Error("output should contain bid price")
	}
	if !strings.Contains(output, "95001") {
		t.Error("output should contain ask price")
	}
}

func TestCmdTickerJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTicker(ctx, mock, []string{"ETH"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"symbol"`) {
		t.Error("JSON should contain symbol field")
	}
	if !strings.Contains(output, `"bid"`) {
		t.Error("JSON should contain bid field")
	}
}

func TestCmdOrderBook(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdOrderBook(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "95001") {
		t.Error("output should contain ask price")
	}
	if !strings.Contains(output, "95000") {
		t.Error("output should contain bid price")
	}
}

func TestCmdOrderBookWithDepth(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		err := cmdOrderBook(ctx, mock, []string{"BTC", "5"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "FetchOrderBook" {
		t.Errorf("expected FetchOrderBook, got %s", mock.lastMethod)
	}
}

func TestCmdOrderBookJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdOrderBook(ctx, mock, []string{"BTC"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"bids"`) {
		t.Error("JSON should contain bids field")
	}
	if !strings.Contains(output, `"asks"`) {
		t.Error("JSON should contain asks field")
	}
}

func TestCmdSymbolDetails(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdSymbolDetails(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "BTC") {
		t.Error("output should contain symbol")
	}
	if !strings.Contains(output, "0.001") {
		t.Error("output should contain min quantity")
	}
}

func TestCmdFeeRate(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdFeeRate(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "0.0002") {
		t.Error("output should contain maker fee")
	}
	if !strings.Contains(output, "0.0005") {
		t.Error("output should contain taker fee")
	}
}

func TestCmdFundingRate(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdFundingRate(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "0.0001") {
		t.Error("output should contain funding rate")
	}
}

func TestCmdFundingRateNonPerp(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdFundingRate(ctx, mock, []string{"BTC"}, false)
	if err == nil {
		t.Error("funding on non-perp should fail")
	}
}

func TestCmdTrades(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTrades(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "FetchTrades" {
		t.Errorf("expected FetchTrades, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "95000") {
		t.Error("output should contain trade price")
	}
}

func TestCmdTradesJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTrades(ctx, mock, []string{"BTC", "5"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"price"`) {
		t.Error("JSON should contain price field")
	}
}

func TestCmdKlines(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdKlines(ctx, mock, []string{"BTC", "1h"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "FetchKlines" {
		t.Errorf("expected FetchKlines, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "94000") {
		t.Error("output should contain open price")
	}
}

func TestCmdKlinesJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdKlines(ctx, mock, []string{"BTC", "1h", "10"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"open"`) {
		t.Error("JSON should contain open field")
	}
}

func TestCmdKlinesNotEnoughArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdKlines(ctx, mock, []string{"BTC"}, false)
	if err == nil {
		t.Error("klines without interval should fail")
	}
}
