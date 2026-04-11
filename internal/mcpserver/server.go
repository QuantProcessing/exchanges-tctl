package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const defaultProtocolVersion = "2025-06-18"

type ToolRunner func(context.Context, ToolCallRequest) (*DecodedCLIResult, error)

type ServerOptions struct {
	Name       string
	Version    string
	ToolRunner ToolRunner
}

type Server struct {
	name       string
	version    string
	toolRunner ToolRunner
}

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type callToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

func NewServer(opts ServerOptions) *Server {
	name := opts.Name
	if name == "" {
		name = "tctl"
	}
	version := opts.Version
	if version == "" {
		version = "dev"
	}
	runner := opts.ToolRunner
	if runner == nil {
		runner = ExecuteTool
	}
	return &Server{
		name:       name,
		version:    version,
		toolRunner: runner,
	}
}

func (s *Server) HandleRequest(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	if req.JSONRPC == "" {
		req.JSONRPC = "2.0"
	}

	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": protocolVersionFromParams(req.Params),
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    s.name,
					"version": s.version,
				},
				"instructions": "tctl MCP exposes request/response market and trading tools only. watch/streaming commands are deferred in v1.",
			},
		}
	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.listTools(),
			},
		}
	case "tools/call":
		return s.handleToolCall(ctx, req)
	case "notifications/initialized":
		return JSONRPCResponse{}
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "method not found",
			},
		}
	}
}

func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			var batch []json.RawMessage
			if err := json.Unmarshal([]byte(line), &batch); err != nil {
				return err
			}

			responses := make([]JSONRPCResponse, 0, len(batch))
			for _, raw := range batch {
				var req JSONRPCRequest
				if err := json.Unmarshal(raw, &req); err != nil {
					responses = append(responses, JSONRPCResponse{
						JSONRPC: "2.0",
						Error: &JSONRPCError{
							Code:    -32700,
							Message: "parse error",
						},
					})
					continue
				}
				resp := s.HandleRequest(ctx, req)
				if resp.JSONRPC == "" {
					continue
				}
				if req.ID == nil {
					continue
				}
				responses = append(responses, resp)
			}
			if len(responses) == 0 {
				continue
			}
			if err := writeJSONLine(out, responses); err != nil {
				return err
			}
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			if err := writeJSONLine(out, JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &JSONRPCError{
					Code:    -32700,
					Message: "parse error",
				},
			}); err != nil {
				return err
			}
			continue
		}

		resp := s.HandleRequest(ctx, req)
		if resp.JSONRPC == "" || req.ID == nil {
			continue
		}
		if err := writeJSONLine(out, resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handleToolCall(ctx context.Context, req JSONRPCRequest) JSONRPCResponse {
	var params callToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "invalid params",
			},
		}
	}

	_, ok := ToolByName(params.Name)
	if !ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32602,
				Message: fmt.Sprintf("unknown tool: %s", params.Name),
			},
		}
	}

	exchange, _ := optionalStringArg(params.Arguments, "exchange")
	market, _ := optionalStringArg(params.Arguments, "market")
	reqArgs := cloneArguments(params.Arguments)
	delete(reqArgs, "exchange")
	delete(reqArgs, "market")

	result, err := s.toolRunner(ctx, ToolCallRequest{
		ToolName:  params.Name,
		Exchange:  exchange,
		Market:    market,
		Arguments: reqArgs,
	})
	if err != nil {
		var cliErr *CLIError
		if errors.As(err, &cliErr) {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  mcpToolErrorResult(cliErr),
			}
		}
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"isError": true,
				"content": []map[string]any{{
					"type": "text",
					"text": err.Error(),
				}},
				"structuredContent": map[string]any{
					"code":    "tool_execution_failed",
					"message": err.Error(),
				},
			},
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []map[string]any{{
				"type": "text",
				"text": string(result.Content),
			}},
			"structuredContent": decodeRawJSON(result.Content),
			"_meta": map[string]any{
				"stderr": result.Metadata.Stderr,
			},
		},
	}
}

func (s *Server) listTools() []map[string]any {
	tools := make([]map[string]any, 0, len(CanonicalToolInventory()))
	for _, tool := range CanonicalToolInventory() {
		tools = append(tools, map[string]any{
			"name":        tool.Name,
			"description": toolDescription(tool),
			"inputSchema": toolInputSchema(tool),
			"annotations": toolAnnotations(tool),
		})
	}
	return tools
}

func protocolVersionFromParams(raw json.RawMessage) string {
	if len(raw) == 0 {
		return defaultProtocolVersion
	}
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return defaultProtocolVersion
	}
	if params.ProtocolVersion == "" {
		return defaultProtocolVersion
	}
	return params.ProtocolVersion
}

func toolDescription(tool ToolDefinition) string {
	switch tool.Name {
	case "execute_funding_arbitrage":
		return "Execute the existing fund-arb composite command using current tctl semantics."
	}
	if tool.Group == ToolGroupMutating {
		return "Mutating tctl command with real side effects."
	}
	return "Read-only request/response tctl command."
}

func toolAnnotations(tool ToolDefinition) map[string]any {
	if tool.Group == ToolGroupMutating {
		return map[string]any{
			"destructiveHint": true,
			"idempotentHint":  false,
			"title":           tool.Name,
		}
	}
	return map[string]any{
		"readOnlyHint": true,
		"title":        tool.Name,
	}
}

func toolInputSchema(tool ToolDefinition) map[string]any {
	properties := map[string]any{
		"exchange": map[string]any{
			"type":        "string",
			"description": "Optional exchange name such as BINANCE or OKX.",
		},
		"market": map[string]any{
			"type":        "string",
			"description": "Optional market type such as perp or spot.",
			"enum":        []string{"perp", "spot"},
		},
	}
	required := []string{}

	addString := func(name, description string) {
		properties[name] = map[string]any{
			"type":        "string",
			"description": description,
		}
	}
	addBool := func(name, description string) {
		properties[name] = map[string]any{
			"type":        "boolean",
			"description": description,
		}
	}
	addInt := func(name, description string) {
		properties[name] = map[string]any{
			"type":        "integer",
			"description": description,
		}
	}

	switch tool.Command {
	case "ticker", "orderbook", "details", "fee", "funding", "trades", "klines":
		addString("symbol", "Trading symbol.")
		required = append(required, "symbol")
	case "order", "cancel":
		addString("order_id", "Order identifier.")
		addString("symbol", "Trading symbol.")
		required = append(required, "order_id", "symbol")
	case "orders", "balance", "account", "positions", "funding-all", "spot-balances":
	case "buy", "sell":
		addString("symbol", "Trading symbol.")
		addString("quantity", "Order quantity.")
		addString("price", "Optional limit price.")
		addString("tif", "Optional time-in-force.")
		addBool("post_only", "Whether the order is post-only.")
		addBool("reduce_only", "Whether the order is reduce-only.")
		addString("client_id", "Optional client order id.")
		required = append(required, "symbol", "quantity")
	case "cancel-all":
		addString("symbol", "Trading symbol.")
		required = append(required, "symbol")
	case "modify":
		addString("order_id", "Order identifier.")
		addString("symbol", "Trading symbol.")
		addString("price", "Optional new price.")
		addString("quantity", "Optional new quantity.")
		required = append(required, "order_id", "symbol")
	case "leverage":
		addString("symbol", "Trading symbol.")
		addString("leverage", "Target leverage.")
		required = append(required, "symbol", "leverage")
	case "transfer":
		addString("asset", "Asset to transfer.")
		addString("amount", "Transfer amount.")
		addString("from", "Source account.")
		addString("to", "Destination account.")
		required = append(required, "asset", "amount")
	case "fund-arb":
		addString("symbol", "Trading symbol.")
		addString("quantity", "Arbitrage quantity.")
		addString("spot_exchange", "Optional spot exchange.")
		addString("perp_exchange", "Optional perp exchange.")
		addInt("leverage", "Optional perp leverage.")
		addBool("close", "Whether to close an existing arbitrage position.")
		addString("spot_price", "Optional spot limit price.")
		addString("perp_price", "Optional perp limit price.")
		required = append(required, "symbol", "quantity")
	}

	if tool.Command == "orderbook" {
		addInt("depth", "Optional order book depth.")
	}
	if tool.Command == "trades" || tool.Command == "klines" {
		addInt("limit", "Optional result limit.")
	}
	if tool.Command == "klines" {
		addString("interval", "Candlestick interval.")
		required = append(required, "interval")
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func mcpToolErrorResult(err *CLIError) map[string]any {
	structured := map[string]any{
		"code":    err.Code,
		"message": err.Message,
	}
	if len(err.StructuredBody) > 0 {
		structured["body"] = decodeRawJSON(err.StructuredBody)
	}
	if err.Metadata.Stdout != "" || err.Metadata.Stderr != "" || err.Metadata.Duration != "" {
		structured["metadata"] = map[string]any{
			"stdout":   err.Metadata.Stdout,
			"stderr":   err.Metadata.Stderr,
			"duration": err.Metadata.Duration,
		}
	}
	return map[string]any{
		"isError": true,
		"content": []map[string]any{{
			"type": "text",
			"text": err.Message,
		}},
		"structuredContent": structured,
	}
}

func decodeRawJSON(raw json.RawMessage) any {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return string(raw)
	}
	return decoded
}

func writeJSONLine(w io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s\n", data); err != nil {
		return err
	}
	return nil
}

func cloneArguments(args map[string]any) map[string]any {
	if len(args) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(args))
	for k, v := range args {
		cloned[k] = v
	}
	return cloned
}
