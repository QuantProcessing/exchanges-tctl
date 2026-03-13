package main

import (
	"context"
	"strings"
	"testing"
)

func TestCmdPositions(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPositions(ctx, mock, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "BTC") {
		t.Error("output should contain BTC position")
	}
	if !strings.Contains(output, "ETH") {
		t.Error("output should contain ETH position")
	}
}

func TestCmdPositionsJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdPositions(ctx, mock, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"symbol"`) {
		t.Error("JSON should contain symbol field")
	}
}

func TestCmdPositionsNonPerp(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdPositions(ctx, mock, nil, false)
	if err == nil {
		t.Error("positions on non-perp should fail")
	}
}

func TestCmdOpenOrders(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdOpenOrders(ctx, mock, nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "o1") {
		t.Error("output should contain order ID")
	}
}

func TestCmdOpenOrdersJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdOpenOrders(ctx, mock, []string{"BTC"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"order_id"`) {
		t.Error("JSON should contain order_id field")
	}
}

func TestCmdBalance(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdBalance(ctx, mock, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10000") {
		t.Error("output should contain balance value")
	}
}

func TestCmdBalanceJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdBalance(ctx, mock, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"10000"`) {
		t.Error("JSON should contain balance value")
	}
}

func TestCmdAccount(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdAccount(ctx, mock, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "10000") {
		t.Error("output should contain total balance")
	}
	if !strings.Contains(output, "8500") {
		t.Error("output should contain available balance")
	}
}

func TestCmdAccountJSON(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdAccount(ctx, mock, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"total_balance"`) {
		t.Error("JSON should contain total_balance field")
	}
}

func TestCmdSetLeverage(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdSetLeverage(ctx, mock, []string{"BTC", "10"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "SetLeverage" {
		t.Errorf("expected SetLeverage, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "10x") {
		t.Error("output should contain leverage value")
	}
}

func TestCmdSetLeverageNonPerp(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdSetLeverage(ctx, mock, []string{"BTC", "10"}, false)
	if err == nil {
		t.Error("leverage on non-perp should fail")
	}
}

func TestCmdSetLeverageInvalidValue(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	err := cmdSetLeverage(ctx, mock, []string{"BTC", "abc"}, false)
	if err == nil {
		t.Error("should error on invalid leverage")
	}
}

// ============================================================================
// Spot-specific commands
// ============================================================================

func TestCmdSpotBalances(t *testing.T) {
	mock := newSpotMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdSpotBalances(ctx, mock, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "BTC") {
		t.Error("output should contain BTC")
	}
	if !strings.Contains(output, "USDT") {
		t.Error("output should contain USDT")
	}
	if !strings.Contains(output, "50000") {
		t.Error("output should contain USDT balance")
	}
}

func TestCmdSpotBalancesJSON(t *testing.T) {
	mock := newSpotMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdSpotBalances(ctx, mock, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"asset"`) {
		t.Error("JSON should contain asset field")
	}
}

func TestCmdSpotBalancesNonSpot(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdSpotBalances(ctx, mock, false)
	if err == nil {
		t.Error("spot-balances on non-spot should fail")
	}
	if !strings.Contains(err.Error(), "spot") {
		t.Errorf("error should mention spot, got: %v", err)
	}
}

func TestCmdTransfer(t *testing.T) {
	mock := newSpotMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTransfer(ctx, mock, []string{"USDT", "1000", "--from", "SPOT", "--to", "PERP"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if mock.lastMethod != "TransferAsset" {
		t.Errorf("expected TransferAsset, got %s", mock.lastMethod)
	}
	if !strings.Contains(output, "Transferred") {
		t.Error("output should confirm transfer")
	}
}

func TestCmdTransferJSON(t *testing.T) {
	mock := newSpotMock()
	ctx := context.Background()

	output := captureOutput(func() {
		err := cmdTransfer(ctx, mock, []string{"BTC", "0.5"}, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, `"status"`) {
		t.Error("JSON should contain status field")
	}
}

func TestCmdTransferNonSpot(t *testing.T) {
	mock := &mockExchange{}
	ctx := context.Background()

	err := cmdTransfer(ctx, mock, []string{"USDT", "100"}, false)
	if err == nil {
		t.Error("transfer on non-spot should fail")
	}
}

func TestCmdTransferInvalidAmount(t *testing.T) {
	mock := newSpotMock()
	ctx := context.Background()

	err := cmdTransfer(ctx, mock, []string{"USDT", "abc"}, false)
	if err == nil {
		t.Error("should error on invalid amount")
	}
}
