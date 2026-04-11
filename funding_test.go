package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

// spotPerpMock implements both spot and perp operations for fund-arb testing.
type spotPerpMock struct {
	exchanges.Exchange
	marketType exchanges.MarketType
	exchange   string
	lastMethod string
	orders     []*exchanges.Order // records all placed orders
	leverages  []setLeverageCall
	failSpot   bool
	failPerp   bool
}

type setLeverageCall struct {
	symbol   string
	leverage int
}

func (m *spotPerpMock) GetExchange() string                 { return m.exchange }
func (m *spotPerpMock) GetMarketType() exchanges.MarketType { return m.marketType }
func (m *spotPerpMock) FormatSymbol(s string) string        { return s }
func (m *spotPerpMock) ExtractSymbol(s string) string       { return s }
func (m *spotPerpMock) Close() error                        { return nil }

func (m *spotPerpMock) PlaceOrder(ctx context.Context, params *exchanges.OrderParams) (*exchanges.Order, error) {
	if m.failSpot && m.marketType == exchanges.MarketTypeSpot {
		return nil, fmt.Errorf("spot order failed")
	}
	if m.failPerp && m.marketType == exchanges.MarketTypePerp {
		return nil, fmt.Errorf("perp order failed")
	}
	order := &exchanges.Order{
		OrderID:  fmt.Sprintf("order-%s-%s", m.marketType, params.Side),
		Symbol:   params.Symbol,
		Side:     params.Side,
		Type:     params.Type,
		Price:    params.Price,
		Quantity: params.Quantity,
		Status:   exchanges.OrderStatusNew,
	}
	m.orders = append(m.orders, order)
	return order, nil
}

func (m *spotPerpMock) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	m.leverages = append(m.leverages, setLeverageCall{symbol, leverage})
	return nil
}

func (m *spotPerpMock) FetchPositions(ctx context.Context) ([]exchanges.Position, error) {
	return nil, nil
}

func (m *spotPerpMock) FetchFundingRate(ctx context.Context, symbol string) (*exchanges.FundingRate, error) {
	return &exchanges.FundingRate{}, nil
}

func (m *spotPerpMock) FetchAllFundingRates(ctx context.Context) ([]exchanges.FundingRate, error) {
	return nil, nil
}

func (m *spotPerpMock) ModifyOrder(ctx context.Context, orderID, symbol string, params *exchanges.ModifyOrderParams) (*exchanges.Order, error) {
	return nil, nil
}

// ============================================================================
// Tests
// ============================================================================

func TestParseNamedPrice(t *testing.T) {
	tests := []struct {
		args     []string
		flag     string
		wantVal  string
		wantRest int
	}{
		{[]string{"--spot-price", "95000"}, "--spot-price", "95000", 0},
		{[]string{"--leverage", "5", "--spot-price", "95000"}, "--spot-price", "95000", 2},
		{[]string{"--leverage", "5"}, "--spot-price", "0", 2},
		{nil, "--spot-price", "0", 0},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, "_"), func(t *testing.T) {
			val, rest := parseNamedPrice(tt.args, tt.flag)
			if val.String() != tt.wantVal {
				t.Errorf("parseNamedPrice value = %s, want %s", val.String(), tt.wantVal)
			}
			if len(rest) != tt.wantRest {
				t.Errorf("remaining args count = %d, want %d", len(rest), tt.wantRest)
			}
		})
	}
}

func TestFundArbArgsValidation(t *testing.T) {
	// Not enough args
	err := cmdFundArb(context.Background(), "MOCK", nil, false, nil)
	if err == nil {
		t.Error("expected error for no args")
	}
	if !strings.Contains(err.Error(), "not enough arguments") {
		t.Errorf("expected 'not enough arguments', got: %v", err)
	}
}

func TestFundArbInvalidQuantity(t *testing.T) {
	err := cmdFundArb(context.Background(), "MOCK", []string{"BTC", "invalid"}, false, nil)
	if err == nil {
		t.Error("expected error for invalid quantity")
	}
	if !strings.Contains(err.Error(), "invalid quantity") {
		t.Errorf("expected 'invalid quantity', got: %v", err)
	}
}

func TestFundArbInvalidLeverage(t *testing.T) {
	err := cmdFundArb(context.Background(), "MOCK", []string{"BTC", "0.01", "--leverage", "abc"}, false, nil)
	if err == nil {
		t.Error("expected error for invalid leverage")
	}
	if !strings.Contains(err.Error(), "invalid leverage") {
		t.Errorf("expected 'invalid leverage', got: %v", err)
	}
}

func TestFundArbCrossExchangeRequiresBothFlags(t *testing.T) {
	err := cmdFundArb(context.Background(), "MOCK", []string{"BTC", "0.01", "--spot-exchange", "BINANCE"}, false, nil)
	if err == nil {
		t.Error("expected error when only one cross-exchange flag is set")
	}
	if !strings.Contains(err.Error(), "must set both --spot-exchange and --perp-exchange") {
		t.Errorf("expected paired-flag validation error, got: %v", err)
	}
}

func TestFundArbCrossExchangeUsesExplicitLegExchanges(t *testing.T) {
	err := cmdFundArb(context.Background(), "", []string{"BTC", "0.01", "--spot-exchange", "SPOTX", "--perp-exchange", "PERPX"}, false, nil)
	if err == nil {
		t.Error("expected error for unsupported spot exchange")
	}
	if !strings.Contains(err.Error(), "unsupported exchange: SPOTX") {
		t.Errorf("expected spot leg exchange to be used, got: %v", err)
	}
}

func TestFundArbWithoutCrossFlagsRequiresResolvableDefaultExchange(t *testing.T) {
	err := cmdFundArb(context.Background(), "", []string{"BTC", "0.01"}, false, nil)
	if err == nil {
		t.Error("expected error when no default exchange can be resolved")
	}
	if !strings.Contains(err.Error(), "no exchange specified and none auto-detected") {
		t.Errorf("expected missing-default-exchange error, got: %v", err)
	}
}

func TestResolveFundArbExchangesDefaultsToSingleExchange(t *testing.T) {
	spotExchange, perpExchange, err := resolveFundArbExchanges("binance", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spotExchange != "BINANCE" || perpExchange != "BINANCE" {
		t.Fatalf("got spot=%q perp=%q, want BINANCE/BINANCE", spotExchange, perpExchange)
	}
}

func TestResolveFundArbExchangesUsesExplicitLegs(t *testing.T) {
	spotExchange, perpExchange, err := resolveFundArbExchanges("binance", "okx", "bybit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spotExchange != "OKX" || perpExchange != "BYBIT" {
		t.Fatalf("got spot=%q perp=%q, want OKX/BYBIT", spotExchange, perpExchange)
	}
}

func TestOutputFundArbTable(t *testing.T) {
	spotOrder := &exchanges.Order{
		OrderID:  "spot-123",
		Symbol:   "BTC",
		Side:     exchanges.OrderSideBuy,
		Quantity: decimal.RequireFromString("0.01"),
		Price:    decimal.Zero,
	}
	perpOrder := &exchanges.Order{
		OrderID:  "perp-456",
		Symbol:   "BTC",
		Side:     exchanges.OrderSideSell,
		Quantity: decimal.RequireFromString("0.01"),
		Price:    decimal.Zero,
	}

	output := captureOutput(func() {
		err := outputFundArbTable("BINANCE", spotOrder, nil, "OKX", perpOrder, nil, false)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "OPEN") {
		t.Error("expected OPEN in output")
	}
	if !strings.Contains(output, "spot-123") {
		t.Error("expected spot order ID in output")
	}
	if !strings.Contains(output, "perp-456") {
		t.Error("expected perp order ID in output")
	}
	if !strings.Contains(output, "BINANCE") || !strings.Contains(output, "OKX") {
		t.Error("expected exchange names in output")
	}
}

func TestOutputFundArbTableClose(t *testing.T) {
	spotOrder := &exchanges.Order{
		OrderID:  "spot-789",
		Symbol:   "BTC",
		Side:     exchanges.OrderSideSell,
		Quantity: decimal.RequireFromString("0.01"),
	}
	perpOrder := &exchanges.Order{
		OrderID:  "perp-012",
		Symbol:   "BTC",
		Side:     exchanges.OrderSideBuy,
		Quantity: decimal.RequireFromString("0.01"),
	}

	output := captureOutput(func() {
		err := outputFundArbTable("BINANCE", spotOrder, nil, "BINANCE", perpOrder, nil, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "CLOSE") {
		t.Error("expected CLOSE in output")
	}
}

func TestOutputFundArbPartialFailure(t *testing.T) {
	spotOrder := &exchanges.Order{
		OrderID:  "spot-ok",
		Symbol:   "BTC",
		Side:     exchanges.OrderSideBuy,
		Quantity: decimal.RequireFromString("0.01"),
	}
	perpErr := fmt.Errorf("insufficient margin")

	var err error
	captureOutput(func() {
		err = outputFundArbTable("BINANCE", spotOrder, nil, "OKX", nil, perpErr, false)
	})

	if err == nil {
		t.Error("expected error for partial failure")
	}
	if !strings.Contains(err.Error(), "perp order failed") {
		t.Errorf("expected 'perp order failed' in error, got: %v", err)
	}
}

func TestOutputFundArbBothFailed(t *testing.T) {
	spotErr := fmt.Errorf("spot error")
	perpErr := fmt.Errorf("perp error")

	var err error
	captureOutput(func() {
		err = outputFundArbTable("BINANCE", nil, spotErr, "OKX", nil, perpErr, false)
	})

	if err == nil {
		t.Error("expected error when both failed")
	}
	if !strings.Contains(err.Error(), "both orders failed") {
		t.Errorf("expected 'both orders failed', got: %v", err)
	}
}

func TestOutputFundArbJSON(t *testing.T) {
	spotOrder := &exchanges.Order{
		OrderID: "spot-json",
		Symbol:  "ETH",
		Side:    exchanges.OrderSideBuy,
	}
	perpOrder := &exchanges.Order{
		OrderID: "perp-json",
		Symbol:  "ETH",
		Side:    exchanges.OrderSideSell,
	}

	output := captureOutput(func() {
		err := outputFundArbJSON("BINANCE", spotOrder, nil, "OKX", perpOrder, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "spot-json") {
		t.Error("expected spot order ID in JSON output")
	}
	if !strings.Contains(output, "perp-json") {
		t.Error("expected perp order ID in JSON output")
	}
	if !strings.Contains(output, `"market": "spot"`) {
		t.Error("expected market:spot in JSON output")
	}
	if !strings.Contains(output, `"exchange": "BINANCE"`) || !strings.Contains(output, `"exchange": "OKX"`) {
		t.Error("expected exchange names in JSON output")
	}
}

func TestOutputFundArbJSONPartialFail(t *testing.T) {
	perpOrder := &exchanges.Order{
		OrderID: "perp-ok",
		Symbol:  "BTC",
	}
	spotErr := fmt.Errorf("no balance")

	output := captureOutput(func() {
		_ = outputFundArbJSON("BINANCE", nil, spotErr, "OKX", perpOrder, nil)
	})

	if !strings.Contains(output, "no balance") {
		t.Error("expected error message in JSON output")
	}
	if !strings.Contains(output, "perp-ok") {
		t.Error("expected perp order in JSON output")
	}
}
