package mcpserver

func CanonicalToolInventory() []ToolDefinition {
	return []ToolDefinition{
		{Name: "get_ticker", Command: "ticker", Group: ToolGroupRead, Aliases: []string{"t"}},
		{Name: "get_order_book", Command: "orderbook", Group: ToolGroupRead, Aliases: []string{"ob"}},
		{Name: "get_symbol_details", Command: "details", Group: ToolGroupRead},
		{Name: "get_fee_rate", Command: "fee", Group: ToolGroupRead},
		{Name: "get_funding_rate", Command: "funding", Group: ToolGroupRead},
		{Name: "get_recent_trades", Command: "trades", Group: ToolGroupRead},
		{Name: "get_klines", Command: "klines", Group: ToolGroupRead, Aliases: []string{"kl"}},
		{Name: "get_order", Command: "order", Group: ToolGroupRead},
		{Name: "list_open_orders", Command: "orders", Group: ToolGroupRead, Aliases: []string{"o"}},
		{Name: "get_balance", Command: "balance", Group: ToolGroupRead, Aliases: []string{"bal", "b"}},
		{Name: "get_account", Command: "account", Group: ToolGroupRead, Aliases: []string{"acc"}},
		{Name: "list_positions", Command: "positions", Group: ToolGroupRead, Aliases: []string{"pos", "p"}},
		{Name: "list_all_funding_rates", Command: "funding-all", Group: ToolGroupRead},
		{Name: "list_spot_balances", Command: "spot-balances", Group: ToolGroupRead, Aliases: []string{"sb"}},
		{Name: "place_buy_order", Command: "buy", Group: ToolGroupMutating},
		{Name: "place_sell_order", Command: "sell", Group: ToolGroupMutating},
		{Name: "cancel_order", Command: "cancel", Group: ToolGroupMutating},
		{Name: "cancel_all_orders", Command: "cancel-all", Group: ToolGroupMutating},
		{Name: "modify_order", Command: "modify", Group: ToolGroupMutating},
		{Name: "set_leverage", Command: "leverage", Group: ToolGroupMutating, Aliases: []string{"lev"}},
		{Name: "transfer_asset", Command: "transfer", Group: ToolGroupMutating},
		{Name: "execute_funding_arbitrage", Command: "fund-arb", Group: ToolGroupMutating, Aliases: []string{"fa"}},
	}
}

func DeferredCommandInventory() map[string]ToolDefinition {
	reason := "v1 only covers request/response commands"
	return map[string]ToolDefinition{
		"watch-ticker":    {Command: "watch-ticker", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wt"}},
		"watch-ob":        {Command: "watch-ob", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wob"}},
		"watch-orders":    {Command: "watch-orders", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wo"}},
		"watch-trades":    {Command: "watch-trades", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wtr"}},
		"watch-positions": {Command: "watch-positions", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wp"}},
		"watch-klines":    {Command: "watch-klines", Group: ToolGroupStreaming, Deferred: true, DeferredReason: reason, Aliases: []string{"wkl"}},
	}
}

func ToolByName(name string) (ToolDefinition, bool) {
	for _, tool := range CanonicalToolInventory() {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolDefinition{}, false
}
