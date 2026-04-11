package mcpserver

import (
	"slices"
	"testing"
)

func TestCanonicalToolInventoryMatchesPRD(t *testing.T) {
	t.Parallel()

	inventory := CanonicalToolInventory()

	if len(inventory) == 0 {
		t.Fatalf("CanonicalToolInventory() returned no tools")
	}

	expectedIncluded := map[string]toolExpectation{
		"get_ticker":                {Command: "ticker", Group: ToolGroupRead},
		"get_order_book":            {Command: "orderbook", Group: ToolGroupRead},
		"get_symbol_details":        {Command: "details", Group: ToolGroupRead},
		"get_fee_rate":              {Command: "fee", Group: ToolGroupRead},
		"get_funding_rate":          {Command: "funding", Group: ToolGroupRead},
		"get_recent_trades":         {Command: "trades", Group: ToolGroupRead},
		"get_klines":                {Command: "klines", Group: ToolGroupRead},
		"get_order":                 {Command: "order", Group: ToolGroupRead},
		"list_open_orders":          {Command: "orders", Group: ToolGroupRead},
		"get_balance":               {Command: "balance", Group: ToolGroupRead},
		"get_account":               {Command: "account", Group: ToolGroupRead},
		"list_positions":            {Command: "positions", Group: ToolGroupRead},
		"list_all_funding_rates":    {Command: "funding-all", Group: ToolGroupRead},
		"list_spot_balances":        {Command: "spot-balances", Group: ToolGroupRead},
		"place_buy_order":           {Command: "buy", Group: ToolGroupMutating},
		"place_sell_order":          {Command: "sell", Group: ToolGroupMutating},
		"cancel_order":              {Command: "cancel", Group: ToolGroupMutating},
		"cancel_all_orders":         {Command: "cancel-all", Group: ToolGroupMutating},
		"modify_order":              {Command: "modify", Group: ToolGroupMutating},
		"set_leverage":              {Command: "leverage", Group: ToolGroupMutating},
		"transfer_asset":            {Command: "transfer", Group: ToolGroupMutating},
		"execute_funding_arbitrage": {Command: "fund-arb", Group: ToolGroupMutating},
	}

	deferredCommands := map[string]string{
		"watch-ticker":    "v1 only covers request/response commands",
		"watch-ob":        "v1 only covers request/response commands",
		"watch-orders":    "v1 only covers request/response commands",
		"watch-trades":    "v1 only covers request/response commands",
		"watch-positions": "v1 only covers request/response commands",
		"watch-klines":    "v1 only covers request/response commands",
	}

	aliasCommands := []string{
		"t", "ob", "kl", "o", "bal", "b", "acc", "pos", "p", "sb", "lev", "fa",
		"wt", "wob", "wo", "wtr", "wp", "wkl",
	}

	seenCommands := make(map[string]ToolDefinition, len(inventory))
	seenNames := make(map[string]ToolDefinition, len(inventory))

	for _, tool := range inventory {
		if _, ok := seenNames[tool.Name]; ok {
			t.Fatalf("duplicate tool name %q in inventory", tool.Name)
		}
		seenNames[tool.Name] = tool

		if tool.Command != "" {
			if _, ok := seenCommands[tool.Command]; ok {
				t.Fatalf("duplicate canonical command %q in inventory", tool.Command)
			}
			seenCommands[tool.Command] = tool
		}
	}

	for name, want := range expectedIncluded {
		tool, ok := seenNames[name]
		if !ok {
			t.Fatalf("missing included tool %q", name)
		}
		if tool.Command != want.Command {
			t.Fatalf("tool %q command = %q, want %q", name, tool.Command, want.Command)
		}
		if tool.Group != want.Group {
			t.Fatalf("tool %q group = %q, want %q", name, tool.Group, want.Group)
		}
		if tool.Deferred {
			t.Fatalf("tool %q must not be marked deferred", name)
		}
	}

	fundArb := seenNames["execute_funding_arbitrage"]
	if fundArb.Command != "fund-arb" {
		t.Fatalf("fund-arb tool command = %q, want fund-arb", fundArb.Command)
	}
	if fundArb.Group != ToolGroupMutating {
		t.Fatalf("fund-arb tool group = %q, want %q", fundArb.Group, ToolGroupMutating)
	}

	for command, wantReason := range deferredCommands {
		tool, ok := DeferredCommandInventory()[command]
		if !ok {
			t.Fatalf("missing deferred command record for %q", command)
		}
		if !tool.Deferred {
			t.Fatalf("deferred command %q must be marked deferred", command)
		}
		if tool.DeferredReason != wantReason {
			t.Fatalf("deferred command %q reason = %q, want %q", command, tool.DeferredReason, wantReason)
		}
	}

	for _, alias := range aliasCommands {
		if _, ok := seenCommands[alias]; ok {
			t.Fatalf("alias %q must not appear as a standalone canonical command", alias)
		}
		if _, ok := seenNames[alias]; ok {
			t.Fatalf("alias %q must not appear as a standalone tool name", alias)
		}
	}

	for _, tool := range inventory {
		if tool.Deferred {
			t.Fatalf("included inventory unexpectedly contains deferred tool %q", tool.Name)
		}
		if slices.Contains([]string{
			"watch-ticker", "watch-ob", "watch-orders", "watch-trades", "watch-positions", "watch-klines",
		}, tool.Command) {
			t.Fatalf("watch/streaming command %q must not be in v1 inventory", tool.Command)
		}
	}
}

type toolExpectation struct {
	Command string
	Group   ToolGroup
}
