package main

import (
	"slices"
	"testing"
)

func TestConfiguredExchangesUsesPrefixlessEnvNames(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "binance-key")
	t.Setenv("HYPERLIQUID_PRIVATE_KEY", "hl-key")

	configured := configuredExchanges()

	if !slices.Contains(configured, "BINANCE") {
		t.Fatalf("configuredExchanges() = %v, want BINANCE", configured)
	}
	if !slices.Contains(configured, "HYPERLIQUID") {
		t.Fatalf("configuredExchanges() = %v, want HYPERLIQUID", configured)
	}
}

func TestResolveExchangeAutoDetectsPrefixlessEnvName(t *testing.T) {
	t.Setenv("OKX_API_KEY", "okx-key")

	if got := resolveExchange(""); got != "OKX" {
		t.Fatalf("resolveExchange(\"\") = %q, want %q", got, "OKX")
	}
}
