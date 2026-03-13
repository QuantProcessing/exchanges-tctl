package main

import (
	"context"
	"strings"
	"testing"

	exchanges "github.com/QuantProcessing/exchanges"
)

func TestCmdPlaceOrderLimit(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.5", "--price", "95000"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "PlaceOrder" {
		t.Errorf("expected PlaceOrder, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "Order placed") {
		t.Error("output should contain success message")
	}
	if !strings.Contains(output, "order-123") {
		t.Error("output should contain order ID")
	}
}

func TestCmdPlaceOrderMarket(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideSell, []string{"ETH", "1"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "MARKET") {
		t.Error("market order should show MARKET in output")
	}
}

func TestCmdPlaceOrderJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.1", "--price", "50000"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"order_id"`) {
		t.Error("JSON output should contain order_id field")
	}
	if !strings.Contains(output, `"order-123"`) {
		t.Error("JSON output should contain the order ID value")
	}
}

func TestCmdPlaceOrderInvalidQty(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "abc"}, false)
	if err == nil {
		t.Error("should error on invalid quantity")
	}
	if !strings.Contains(err.Error(), "invalid quantity") {
		t.Errorf("error should mention 'invalid quantity', got: %v", err)
	}
}

func TestCmdPlaceOrderNotEnoughArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC"}, false)
	if err == nil {
		t.Error("should error on not enough args")
	}
}

func TestCmdCancelOrder(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdCancelOrder(ctx, mock, []string{"order-123", "BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "CancelOrder" {
		t.Errorf("expected CancelOrder, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "cancelled") {
		t.Error("output should mention cancelled")
	}
}

func TestCmdCancelOrderJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdCancelOrder(ctx, mock, []string{"order-123", "BTC"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"cancelled"`) {
		t.Error("JSON should contain cancelled status")
	}
}

func TestCmdCancelAllOrders(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdCancelAllOrders(ctx, mock, []string{"BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "cancelled") {
		t.Error("output should confirm cancellation")
	}
}

func TestCmdModifyOrder(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdModifyOrder(ctx, mock, []string{"order-1", "BTC", "--price", "96000", "--qty", "0.3"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "ModifyOrder" {
		t.Errorf("expected ModifyOrder, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "modified") {
		t.Error("output should mention modified")
	}
}

func TestCmdModifyOrderNonPerp(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdModifyOrder(ctx, mock, []string{"order-1", "BTC"}, false)
	if err == nil {
		t.Error("modify on non-perp should fail")
	}
	if !strings.Contains(err.Error(), "perpetual") {
		t.Errorf("error should mention perpetual, got: %v", err)
	}
}

func TestCmdPlaceOrderWithReduceOnly(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideSell, []string{"BTC", "0.5", "--reduce-only"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "PlaceOrder" {
		t.Errorf("expected PlaceOrder, got %s", mock.lastMethod)
	}
}

func TestCmdFetchOrder(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdFetchOrder(ctx, mock, []string{"order-123", "BTC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "FetchOrder" {
		t.Errorf("expected FetchOrder, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "order-123") {
		t.Error("output should contain order ID")
	}
}

func TestCmdFetchOrderJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdFetchOrder(ctx, mock, []string{"order-123", "BTC"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"order_id"`) {
		t.Error("JSON should contain order_id field")
	}
}

func TestCmdFetchOrderMissingArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdFetchOrder(ctx, mock, []string{"order-123"}, false)
	if err == nil {
		t.Error("should error with only 1 arg")
	}
}

func TestCmdPlaceOrderWithTIF(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.1", "--price", "50000", "--tif", "IOC"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "PlaceOrder" {
		t.Errorf("expected PlaceOrder, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "Order placed") {
		t.Error("output should show success")
	}
}

func TestCmdPlaceOrderWithPostOnly(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.1", "--price", "50000", "--post-only"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "PlaceOrder" {
		t.Errorf("expected PlaceOrder, got %s", mock.lastMethod)
	}
}

func TestCmdPlaceOrderInvalidTIF(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.1", "--price", "50000", "--tif", "INVALID"}, false)
	if err == nil {
		t.Error("should error on invalid TIF")
	}
	if !strings.Contains(err.Error(), "invalid time-in-force") {
		t.Errorf("error should mention time-in-force, got: %v", err)
	}
}

func TestCmdPlaceOrderWithClientID(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	captureOutput(func() {
		err := cmdPlaceOrder(ctx, mock, exchanges.OrderSideBuy, []string{"BTC", "0.1", "--price", "50000", "--client-id", "my-order-01"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "PlaceOrder" {
		t.Errorf("expected PlaceOrder, got %s", mock.lastMethod)
	}
}
