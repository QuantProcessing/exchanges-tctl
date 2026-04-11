package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestServerInitializeAndListTools(t *testing.T) {
	t.Parallel()

	server := NewServer(ServerOptions{
		Name:    "tctl",
		Version: "dev",
	})

	initResp := server.HandleRequest(context.Background(), JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "initialize",
		Params: mustJSON(t, map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		}),
	})

	if initResp.Error != nil {
		t.Fatalf("initialize returned error: %+v", initResp.Error)
	}

	initResult := decodeMapResult(t, initResp.Result)
	if initResult["protocolVersion"] != "2025-06-18" {
		t.Fatalf("protocolVersion = %#v, want 2025-06-18", initResult["protocolVersion"])
	}

	serverInfo := initResult["serverInfo"].(map[string]any)
	if serverInfo["name"] != "tctl" {
		t.Fatalf("serverInfo.name = %#v, want tctl", serverInfo["name"])
	}

	capabilities := initResult["capabilities"].(map[string]any)
	if _, ok := capabilities["tools"]; !ok {
		t.Fatalf("initialize capabilities missing tools: %#v", capabilities)
	}

	listResp := server.HandleRequest(context.Background(), JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(2),
		Method:  "tools/list",
	})
	if listResp.Error != nil {
		t.Fatalf("tools/list returned error: %+v", listResp.Error)
	}

	listResult := decodeMapResult(t, listResp.Result)
	rawTools, ok := listResult["tools"].([]any)
	if !ok {
		t.Fatalf("tools/list tools = %#v", listResult["tools"])
	}
	if len(rawTools) != len(CanonicalToolInventory()) {
		t.Fatalf("tools/list count = %d, want %d", len(rawTools), len(CanonicalToolInventory()))
	}
}

func TestServerCallToolMapsSuccessAndToolError(t *testing.T) {
	t.Parallel()

	server := NewServer(ServerOptions{
		Name:    "tctl",
		Version: "dev",
		ToolRunner: func(_ context.Context, req ToolCallRequest) (*DecodedCLIResult, error) {
			switch req.ToolName {
			case "get_ticker":
				return &DecodedCLIResult{
					Success: true,
					Content: json.RawMessage(`{"symbol":"SOLUSDT","price":"123.45"}`),
				}, nil
			case "place_buy_order":
				return nil, &CLIError{
					Code:           "cli_command_failed",
					Message:        "insufficient balance",
					StructuredBody: json.RawMessage(`{"error":"insufficient balance"}`),
					Metadata: CLIResultMetadata{
						Stderr: "exchange rejected order",
					},
				}
			default:
				return nil, errors.New("unexpected tool")
			}
		},
	})

	successResp := server.HandleRequest(context.Background(), JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "success",
		Method:  "tools/call",
		Params: mustJSON(t, map[string]any{
			"name": "get_ticker",
			"arguments": map[string]any{
				"symbol": "SOLUSDT",
			},
		}),
	})
	if successResp.Error != nil {
		t.Fatalf("tools/call success returned protocol error: %+v", successResp.Error)
	}
	successResult := decodeMapResult(t, successResp.Result)
	if successResult["isError"] != nil {
		t.Fatalf("tools/call success isError = %#v, want nil/false", successResult["isError"])
	}

	structured, ok := successResult["structuredContent"].(map[string]any)
	if !ok {
		t.Fatalf("tools/call success structuredContent = %#v", successResult["structuredContent"])
	}
	if structured["symbol"] != "SOLUSDT" {
		t.Fatalf("structuredContent.symbol = %#v, want SOLUSDT", structured["symbol"])
	}

	errorResp := server.HandleRequest(context.Background(), JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "failure",
		Method:  "tools/call",
		Params: mustJSON(t, map[string]any{
			"name": "place_buy_order",
			"arguments": map[string]any{
				"symbol":   "BTCUSDT",
				"quantity": "0.1",
			},
		}),
	})
	if errorResp.Error != nil {
		t.Fatalf("tools/call tool failure returned protocol error: %+v", errorResp.Error)
	}
	errorResult := decodeMapResult(t, errorResp.Result)
	if errorResult["isError"] != true {
		t.Fatalf("tools/call tool failure isError = %#v, want true", errorResult["isError"])
	}

	errStructured := errorResult["structuredContent"].(map[string]any)
	if errStructured["code"] != "cli_command_failed" {
		t.Fatalf("structuredContent.code = %#v, want cli_command_failed", errStructured["code"])
	}
	if errStructured["message"] != "insufficient balance" {
		t.Fatalf("structuredContent.message = %#v", errStructured["message"])
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}

func decodeMapResult(t *testing.T, value any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal(result) error = %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("json.Unmarshal(result) error = %v", err)
	}
	return out
}
