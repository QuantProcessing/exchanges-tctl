package main

import "testing"

func TestIsMCPServeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want bool
	}{
		{args: []string{"mcp", "serve"}, want: true},
		{args: []string{"MCP", "SERVE"}, want: true},
		{args: []string{"ticker", "BTC"}, want: false},
		{args: []string{"mcp"}, want: false},
		{args: nil, want: false},
	}

	for _, tt := range tests {
		if got := isMCPServeCommand(tt.args); got != tt.want {
			t.Fatalf("isMCPServeCommand(%v) = %v, want %v", tt.args, got, tt.want)
		}
	}
}
