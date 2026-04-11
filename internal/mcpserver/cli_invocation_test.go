package mcpserver

import (
	"reflect"
	"testing"
)

func TestBuildCLIInvocationForRepresentativeTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  ToolCallRequest
		want []string
	}{
		{
			name: "read tool injects global flags and json",
			req: ToolCallRequest{
				ToolName: "get_ticker",
				Exchange: "BINANCE",
				Market:   "spot",
				Arguments: map[string]any{
					"symbol": "SOLUSDT",
				},
			},
			want: []string{"-e", "BINANCE", "-m", "spot", "-json", "ticker", "SOLUSDT"},
		},
		{
			name: "mutating tool injects json and forwards optional order flags",
			req: ToolCallRequest{
				ToolName: "place_buy_order",
				Exchange: "BINANCE",
				Market:   "spot",
				Arguments: map[string]any{
					"symbol":    "BTCUSDT",
					"quantity":  "0.1",
					"price":     "95000",
					"post_only": true,
					"client_id": "codex-order-1",
				},
			},
			want: []string{
				"-e", "BINANCE",
				"-m", "spot",
				"-json",
				"buy", "BTCUSDT", "0.1",
				"--price", "95000",
				"--post-only",
				"--client-id", "codex-order-1",
			},
		},
		{
			name: "fund arb preserves composite command and dedicated flags",
			req: ToolCallRequest{
				ToolName: "execute_funding_arbitrage",
				Exchange: "BINANCE",
				Arguments: map[string]any{
					"symbol":        "BTC",
					"quantity":      "0.01",
					"spot_exchange": "BINANCE",
					"perp_exchange": "OKX",
					"leverage":      5,
					"close":         true,
					"spot_price":    "95000",
					"perp_price":    "95100",
				},
			},
			want: []string{
				"-e", "BINANCE",
				"-json",
				"fund-arb", "BTC", "0.01",
				"--spot-exchange", "BINANCE",
				"--perp-exchange", "OKX",
				"--leverage", "5",
				"--close",
				"--spot-price", "95000",
				"--perp-price", "95100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			invocation, err := BuildCLIInvocation(tt.req)
			if err != nil {
				t.Fatalf("BuildCLIInvocation() error = %v", err)
			}

			if invocation.Command == "" {
				t.Fatalf("BuildCLIInvocation() returned empty binary path")
			}

			if !reflect.DeepEqual(invocation.Args, tt.want) {
				t.Fatalf("BuildCLIInvocation() args = %#v, want %#v", invocation.Args, tt.want)
			}

			for _, arg := range invocation.Args {
				if arg == "-ws" || arg == "--ws" {
					t.Fatalf("BuildCLIInvocation() must not inject websocket flag, args = %#v", invocation.Args)
				}
			}
		})
	}
}
