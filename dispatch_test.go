package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	exchanges "github.com/QuantProcessing/exchanges"

	"github.com/shopspring/decimal"
)

// ============================================================================
// Full Mock Adapter — implements Exchange, PerpExchange, and SpotExchange
// ============================================================================

type fullMockAdapter struct {
	exchanges.Exchange // embed to satisfy interface
	lastMethod       string
	marketType       exchanges.MarketType
	exchange         string

	// Configurable return data
	returnError bool
}

func newPerpMock() *fullMockAdapter {
	return &fullMockAdapter{marketType: exchanges.MarketTypePerp, exchange: "MOCK"}
}

func newSpotMock() *fullMockAdapter {
	return &fullMockAdapter{marketType: exchanges.MarketTypeSpot, exchange: "MOCK"}
}

// mockExchange is a lightweight mock that only implements exchanges.Exchange
// (not PerpExchange or SpotExchange). Used to test perp/spot-only command errors.
type mockExchange struct {
	exchanges.Exchange // embed to satisfy interface
	lastMethod       string
}

func (m *mockExchange) GetExchange() string               { return "MOCK" }
func (m *mockExchange) GetMarketType() exchanges.MarketType { return exchanges.MarketTypePerp }
func (m *mockExchange) FormatSymbol(s string) string      { return s }
func (m *mockExchange) ExtractSymbol(s string) string     { return s }

func (m *mockExchange) FetchTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	m.lastMethod = "FetchTicker"
	return &exchanges.Ticker{Symbol: symbol, Bid: decimal.NewFromInt(100), Ask: decimal.NewFromInt(101)}, nil
}

func (m *mockExchange) FetchBalance(ctx context.Context) (decimal.Decimal, error) {
	m.lastMethod = "FetchBalance"
	return decimal.NewFromInt(1000), nil
}

func (m *mockExchange) FetchOrderBook(ctx context.Context, symbol string, limit int) (*exchanges.OrderBook, error) {
	m.lastMethod = "FetchOrderBook"
	return &exchanges.OrderBook{Symbol: symbol}, nil
}

func (m *mockExchange) FetchOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	m.lastMethod = "FetchOpenOrders"
	return nil, nil
}

func (m *mockExchange) FetchSymbolDetails(ctx context.Context, symbol string) (*exchanges.SymbolDetails, error) {
	m.lastMethod = "FetchSymbolDetails"
	return &exchanges.SymbolDetails{Symbol: symbol}, nil
}

func (m *mockExchange) FetchFeeRate(ctx context.Context, symbol string) (*exchanges.FeeRate, error) {
	m.lastMethod = "FetchFeeRate"
	return &exchanges.FeeRate{}, nil
}

func (m *mockExchange) FetchAccount(ctx context.Context) (*exchanges.Account, error) {
	m.lastMethod = "FetchAccount"
	return &exchanges.Account{}, nil
}

func (m *fullMockAdapter) GetExchange() string               { return m.exchange }
func (m *fullMockAdapter) GetMarketType() exchanges.MarketType { return m.marketType }
func (m *fullMockAdapter) FormatSymbol(s string) string      { return s }
func (m *fullMockAdapter) ExtractSymbol(s string) string     { return s }
func (m *fullMockAdapter) Close() error                      { return nil }

// --- Market Data ---

func (m *fullMockAdapter) FetchTicker(ctx context.Context, symbol string) (*exchanges.Ticker, error) {
	m.lastMethod = "FetchTicker"
	return &exchanges.Ticker{
		Symbol:    symbol,
		Bid:       decimal.RequireFromString("95000"),
		Ask:       decimal.RequireFromString("95001"),
		LastPrice: decimal.RequireFromString("95000.5"),
		Volume24h: decimal.RequireFromString("12345.6"),
	}, nil
}

func (m *fullMockAdapter) FetchOrderBook(ctx context.Context, symbol string, limit int) (*exchanges.OrderBook, error) {
	m.lastMethod = "FetchOrderBook"
	return &exchanges.OrderBook{
		Symbol: symbol,
		Bids: []exchanges.Level{
			{Price: decimal.RequireFromString("95000"), Quantity: decimal.RequireFromString("1.5")},
			{Price: decimal.RequireFromString("94999"), Quantity: decimal.RequireFromString("2.0")},
		},
		Asks: []exchanges.Level{
			{Price: decimal.RequireFromString("95001"), Quantity: decimal.RequireFromString("0.8")},
			{Price: decimal.RequireFromString("95002"), Quantity: decimal.RequireFromString("1.2")},
		},
	}, nil
}

func (m *fullMockAdapter) FetchTrades(ctx context.Context, symbol string, limit int) ([]exchanges.Trade, error) {
	m.lastMethod = "FetchTrades"
	return []exchanges.Trade{
		{ID: "1", Symbol: symbol, Price: decimal.RequireFromString("95000"), Quantity: decimal.RequireFromString("0.5"), Side: exchanges.TradeSideBuy, Timestamp: time.Now().UnixMilli()},
		{ID: "2", Symbol: symbol, Price: decimal.RequireFromString("95001"), Quantity: decimal.RequireFromString("0.3"), Side: exchanges.TradeSideSell, Timestamp: time.Now().UnixMilli()},
	}, nil
}

func (m *fullMockAdapter) FetchKlines(ctx context.Context, symbol string, interval exchanges.Interval, opts *exchanges.KlineOpts) ([]exchanges.Kline, error) {
	m.lastMethod = "FetchKlines"
	return []exchanges.Kline{
		{Symbol: symbol, Interval: interval, Open: decimal.RequireFromString("94000"), High: decimal.RequireFromString("96000"), Low: decimal.RequireFromString("93500"), Close: decimal.RequireFromString("95000"), Volume: decimal.RequireFromString("100"), Timestamp: time.Now().UnixMilli()},
	}, nil
}

func (m *fullMockAdapter) FetchSymbolDetails(ctx context.Context, symbol string) (*exchanges.SymbolDetails, error) {
	m.lastMethod = "FetchSymbolDetails"
	return &exchanges.SymbolDetails{
		Symbol:            symbol,
		PricePrecision:    2,
		QuantityPrecision: 4,
		MinQuantity:       decimal.RequireFromString("0.001"),
		MinNotional:       decimal.RequireFromString("10"),
	}, nil
}

func (m *fullMockAdapter) FetchFeeRate(ctx context.Context, symbol string) (*exchanges.FeeRate, error) {
	m.lastMethod = "FetchFeeRate"
	return &exchanges.FeeRate{
		Maker: decimal.RequireFromString("0.0002"),
		Taker: decimal.RequireFromString("0.0005"),
	}, nil
}

// --- Trading ---

func (m *fullMockAdapter) PlaceOrder(ctx context.Context, params *exchanges.OrderParams) (*exchanges.Order, error) {
	m.lastMethod = "PlaceOrder"
	return &exchanges.Order{
		OrderID:  "order-123",
		Symbol:   params.Symbol,
		Side:     params.Side,
		Type:     params.Type,
		Price:    params.Price,
		Quantity: params.Quantity,
		Status:   exchanges.OrderStatusNew,
	}, nil
}

func (m *fullMockAdapter) CancelOrder(ctx context.Context, orderID, symbol string) error {
	m.lastMethod = "CancelOrder"
	return nil
}

func (m *fullMockAdapter) CancelAllOrders(ctx context.Context, symbol string) error {
	m.lastMethod = "CancelAllOrders"
	return nil
}

func (m *fullMockAdapter) FetchOrder(ctx context.Context, orderID, symbol string) (*exchanges.Order, error) {
	m.lastMethod = "FetchOrder"
	return &exchanges.Order{OrderID: orderID, Symbol: symbol}, nil
}

func (m *fullMockAdapter) FetchOpenOrders(ctx context.Context, symbol string) ([]exchanges.Order, error) {
	m.lastMethod = "FetchOpenOrders"
	return []exchanges.Order{
		{OrderID: "o1", Symbol: "BTC", Side: exchanges.OrderSideBuy, Type: exchanges.OrderTypeLimit, Price: decimal.RequireFromString("94000"), Quantity: decimal.RequireFromString("0.1"), Status: exchanges.OrderStatusNew},
	}, nil
}

// --- Account ---

func (m *fullMockAdapter) FetchAccount(ctx context.Context) (*exchanges.Account, error) {
	m.lastMethod = "FetchAccount"
	return &exchanges.Account{
		TotalBalance:     decimal.RequireFromString("10000"),
		AvailableBalance: decimal.RequireFromString("8500"),
		UnrealizedPnL:    decimal.RequireFromString("150.5"),
		Positions: []exchanges.Position{
			{Symbol: "BTC", Side: exchanges.PositionSideLong, Quantity: decimal.RequireFromString("0.5"), EntryPrice: decimal.RequireFromString("94000"), UnrealizedPnL: decimal.RequireFromString("150.5")},
		},
	}, nil
}

func (m *fullMockAdapter) FetchBalance(ctx context.Context) (decimal.Decimal, error) {
	m.lastMethod = "FetchBalance"
	return decimal.RequireFromString("10000"), nil
}

// --- PerpExchange ---

func (m *fullMockAdapter) FetchPositions(ctx context.Context) ([]exchanges.Position, error) {
	m.lastMethod = "FetchPositions"
	return []exchanges.Position{
		{Symbol: "BTC", Side: exchanges.PositionSideLong, Quantity: decimal.RequireFromString("0.5"), EntryPrice: decimal.RequireFromString("94000"), UnrealizedPnL: decimal.RequireFromString("500"), Leverage: decimal.RequireFromString("10")},
		{Symbol: "ETH", Side: exchanges.PositionSideShort, Quantity: decimal.RequireFromString("5"), EntryPrice: decimal.RequireFromString("3500"), UnrealizedPnL: decimal.RequireFromString("-50"), Leverage: decimal.RequireFromString("5")},
	}, nil
}

func (m *fullMockAdapter) SetLeverage(ctx context.Context, symbol string, leverage int) error {
	m.lastMethod = "SetLeverage"
	return nil
}

func (m *fullMockAdapter) FetchFundingRate(ctx context.Context, symbol string) (*exchanges.FundingRate, error) {
	m.lastMethod = "FetchFundingRate"
	return &exchanges.FundingRate{
		Symbol:          symbol,
		FundingRate:     decimal.RequireFromString("0.0001"),
		NextFundingTime: time.Now().Add(time.Hour).UnixMilli(),
	}, nil
}

func (m *fullMockAdapter) FetchAllFundingRates(ctx context.Context) ([]exchanges.FundingRate, error) {
	m.lastMethod = "FetchAllFundingRates"
	return nil, nil
}

func (m *fullMockAdapter) ModifyOrder(ctx context.Context, orderID, symbol string, params *exchanges.ModifyOrderParams) (*exchanges.Order, error) {
	m.lastMethod = "ModifyOrder"
	return &exchanges.Order{
		OrderID:  orderID,
		Symbol:   symbol,
		Side:     exchanges.OrderSideBuy,
		Type:     exchanges.OrderTypeLimit,
		Price:    params.Price,
		Quantity: params.Quantity,
		Status:   exchanges.OrderStatusNew,
	}, nil
}

// --- SpotExchange ---

func (m *fullMockAdapter) FetchSpotBalances(ctx context.Context) ([]exchanges.SpotBalance, error) {
	m.lastMethod = "FetchSpotBalances"
	return []exchanges.SpotBalance{
		{Asset: "BTC", Free: decimal.RequireFromString("1.5"), Locked: decimal.RequireFromString("0.3"), Total: decimal.RequireFromString("1.8")},
		{Asset: "USDT", Free: decimal.RequireFromString("50000"), Locked: decimal.RequireFromString("10000"), Total: decimal.RequireFromString("60000")},
	}, nil
}

func (m *fullMockAdapter) TransferAsset(ctx context.Context, params *exchanges.TransferParams) error {
	m.lastMethod = "TransferAsset"
	return nil
}

// --- Streamable (stubs for WSS) ---

func (m *fullMockAdapter) WatchOrders(ctx context.Context, cb exchanges.OrderUpdateCallback) error {
	m.lastMethod = "WatchOrders"
	return nil
}

func (m *fullMockAdapter) WatchPositions(ctx context.Context, cb exchanges.PositionUpdateCallback) error {
	m.lastMethod = "WatchPositions"
	return nil
}

func (m *fullMockAdapter) WatchTicker(ctx context.Context, symbol string, cb exchanges.TickerCallback) error {
	m.lastMethod = "WatchTicker"
	return nil
}

func (m *fullMockAdapter) WatchTrades(ctx context.Context, symbol string, cb exchanges.TradeCallback) error {
	m.lastMethod = "WatchTrades"
	return nil
}

func (m *fullMockAdapter) WatchKlines(ctx context.Context, symbol string, interval exchanges.Interval, cb exchanges.KlineCallback) error {
	m.lastMethod = "WatchKlines"
	return nil
}

func (m *fullMockAdapter) WatchOrderBook(ctx context.Context, symbol string, cb exchanges.OrderBookCallback) error {
	m.lastMethod = "WatchOrderBook"
	return nil
}

func (m *fullMockAdapter) GetLocalOrderBook(symbol string, depth int) *exchanges.OrderBook {
	return nil
}

func (m *fullMockAdapter) StopWatchOrderBook(ctx context.Context, symbol string) error { return nil }
func (m *fullMockAdapter) StopWatchOrders(ctx context.Context) error                   { return nil }
func (m *fullMockAdapter) StopWatchPositions(ctx context.Context) error                { return nil }
func (m *fullMockAdapter) StopWatchTicker(ctx context.Context, symbol string) error    { return nil }
func (m *fullMockAdapter) StopWatchTrades(ctx context.Context, symbol string) error    { return nil }
func (m *fullMockAdapter) StopWatchKlines(ctx context.Context, symbol string, interval exchanges.Interval) error {
	return nil
}

// ============================================================================
// Helper: Capture stdout for testing
// ============================================================================

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// ============================================================================
// Dispatch Routing Tests — Full Coverage
// ============================================================================

func TestDispatchRouting(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	tests := []struct {
		cmd        string
		args       []string
		wantMethod string
	}{
		// Market data
		{"ticker", []string{"BTC"}, "FetchTicker"},
		{"t", []string{"BTC"}, "FetchTicker"},
		{"orderbook", []string{"BTC"}, "FetchOrderBook"},
		{"ob", []string{"BTC"}, "FetchOrderBook"},
		{"details", []string{"BTC"}, "FetchSymbolDetails"},
		{"fee", []string{"BTC"}, "FetchFeeRate"},
		{"trades", []string{"BTC"}, "FetchTrades"},
		{"klines", []string{"BTC", "1h"}, "FetchKlines"},
		{"kl", []string{"BTC", "1h"}, "FetchKlines"},

		// Trading
		{"buy", []string{"BTC", "0.1"}, "PlaceOrder"},
		{"sell", []string{"ETH", "1"}, "PlaceOrder"},
		{"cancel", []string{"order-1", "BTC"}, "CancelOrder"},
		{"cancel-all", []string{"BTC"}, "CancelAllOrders"},
		{"modify", []string{"order-1", "BTC", "--price", "95000"}, "ModifyOrder"},
		{"order", []string{"order-1", "BTC"}, "FetchOrder"},

		// Account
		{"balance", nil, "FetchBalance"},
		{"bal", nil, "FetchBalance"},
		{"b", nil, "FetchBalance"},
		{"orders", nil, "FetchOpenOrders"},
		{"o", nil, "FetchOpenOrders"},
		{"account", nil, "FetchAccount"},
		{"acc", nil, "FetchAccount"},

		// Perp-specific
		{"positions", nil, "FetchPositions"},
		{"pos", nil, "FetchPositions"},
		{"p", nil, "FetchPositions"},
		{"funding", []string{"BTC"}, "FetchFundingRate"},
		{"funding-all", nil, "FetchAllFundingRates"},
		{"leverage", []string{"BTC", "10"}, "SetLeverage"},
		{"lev", []string{"BTC", "5"}, "SetLeverage"},

		// Spot-specific
		{"spot-balances", nil, "FetchSpotBalances"},
		{"sb", nil, "FetchSpotBalances"},
		{"transfer", []string{"USDT", "100"}, "TransferAsset"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			mock.lastMethod = ""
			// Capture output to avoid polluting test output
			captureOutput(func() {
				err := dispatch(ctx, mock, "MOCK", tt.cmd, tt.args, true)
				if err != nil {
					t.Errorf("dispatch(%q) error = %v", tt.cmd, err)
				}
			})
			if mock.lastMethod != tt.wantMethod {
				t.Errorf("dispatch(%q) called %q, want %q", tt.cmd, mock.lastMethod, tt.wantMethod)
			}
		})
	}
}

func TestDispatchUnknownCommand(t *testing.T) {
	mock := newPerpMock()
	err := dispatch(context.Background(), mock, "MOCK", "unknown-cmd", nil, false)
	if err == nil {
		t.Error("dispatch(unknown-cmd) should return error")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error should mention 'unknown command', got: %v", err)
	}
}

func TestDispatchRequiresArgs(t *testing.T) {
	mock := newPerpMock()
	ctx := context.Background()

	tests := []struct {
		cmd     string
		minArgs int
	}{
		{"ticker", 1},
		{"orderbook", 1},
		{"details", 1},
		{"fee", 1},
		{"funding", 1},
		{"buy", 2},
		{"sell", 2},
		{"cancel", 2},
		{"cancel-all", 1},
		{"modify", 2},
		{"leverage", 2},
		{"trades", 1},
		{"klines", 2},
		{"transfer", 2},
		{"order", 2},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_no_args", func(t *testing.T) {
			err := dispatch(ctx, mock, "MOCK", tt.cmd, nil, false)
			if err == nil {
				t.Errorf("dispatch(%q, nil) should fail due to missing args", tt.cmd)
			}
		})
	}
}

// ============================================================================
// Perp-only commands against spot adapter
// ============================================================================

func TestPerpCommandsOnSpotAdapter(t *testing.T) {
	// spotOnly doesn't implement PerpExchange
	spotOnly := &mockExchange{}
	ctx := context.Background()

	perpCmds := []struct {
		cmd  string
		args []string
	}{
		{"positions", nil},
		{"funding", []string{"BTC"}},
		{"leverage", []string{"BTC", "10"}},
		{"modify", []string{"order-1", "BTC", "--price", "95000"}},
	}

	for _, tt := range perpCmds {
		t.Run(tt.cmd+"_on_spot", func(t *testing.T) {
			output := captureOutput(func() {
				err := dispatch(ctx, spotOnly, "MOCK", tt.cmd, tt.args, false)
				if err == nil {
					t.Errorf("dispatch(%q) on spot adapter should return error", tt.cmd)
				}
			})
			_ = output // we check error, not output
		})
	}
}

// ============================================================================
// WSS Streaming dispatch
// ============================================================================

func TestDispatchStreamingCommands(t *testing.T) {
	mock := newPerpMock()

	tests := []struct {
		cmd        string
		args       []string
		wantMethod string
	}{
		{"watch-ticker", []string{"BTC"}, "WatchTicker"},
		{"wt", []string{"BTC"}, "WatchTicker"},
		{"watch-ob", []string{"BTC"}, "WatchOrderBook"},
		{"wob", []string{"BTC"}, "WatchOrderBook"},
		{"watch-orders", nil, "WatchOrders"},
		{"wo", nil, "WatchOrders"},
		{"watch-trades", []string{"BTC"}, "WatchTrades"},
		{"wtr", []string{"BTC"}, "WatchTrades"},
		{"watch-positions", nil, "WatchPositions"},
		{"wp", nil, "WatchPositions"},
		{"watch-klines", []string{"BTC", "1m"}, "WatchKlines"},
		{"wkl", []string{"BTC", "1m"}, "WatchKlines"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			mock.lastMethod = ""
			// Use a pre-cancelled context so streaming commands don't block
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()
			captureOutput(func() {
				dispatch(cancelCtx, mock, "MOCK", tt.cmd, tt.args, true)
			})
			if mock.lastMethod != tt.wantMethod {
				t.Errorf("dispatch(%q) called %q, want %q", tt.cmd, mock.lastMethod, tt.wantMethod)
			}
		})
	}
}
