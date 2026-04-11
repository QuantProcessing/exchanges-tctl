---
name: tctl
description: Use the tctl MCP plugin for request/response tools, with analysis-first read workflows and explicit side-effecting mutating tools for direct execution.
---

# tctl MCP Skill

## Purpose

Use this skill to decide when to inspect state and when to act. The plugin only covers existing `tctl` request/response commands. `watch-*` and other streaming flows are deferred in `v1` and must not be treated as available capabilities.

## Tool Model

- `read` tools query existing state and are the default for exploratory requests.
- `mutating` tools perform real side effects on exchanges or account state and must be treated as execution actions, not suggestions.
- `execute_funding_arbitrage` is a standalone composite mutating tool. Do not rewrite it as a manual sequence of smaller orders.
- Alias names are compatibility knowledge only. Use the canonical tool names exposed by the plugin.

Representative tools:

- Read: `get_ticker`, `get_order_book`, `get_symbol_details`, `get_fee_rate`, `get_funding_rate`, `get_recent_trades`, `get_klines`, `get_order`, `list_open_orders`, `get_balance`, `get_account`, `list_positions`, `list_all_funding_rates`, `list_spot_balances`
- Mutating: `place_buy_order`, `place_sell_order`, `cancel_order`, `cancel_all_orders`, `modify_order`, `set_leverage`, `transfer_asset`, `execute_funding_arbitrage`

## Analysis-Only Workflow

Use `read` tools first when the user is asking to inspect, compare, summarize, or diagnose market or account state. Stay in analysis mode and return a conclusion instead of acting.

Typical requests:

- check funding or basis conditions
- inspect account, balance, positions, orders, or fees
- compare exchange state before deciding whether to trade

## Direct Execution Workflow

Use `mutating` tools only when the user gives a direct action request, or when the analysis result has already been explicitly confirmed and the next step is execution.

Typical requests:

- place or cancel orders
- modify open orders
- change leverage
- transfer assets
- execute funding arbitrage

## Workflow

1. Classify the request as analysis-only or direct execution.
2. If the request is analysis-only, use `read` tools and stop at the conclusion.
3. If the request is executable, call the matching `mutating` tool directly.
4. For funding arbitrage, use `execute_funding_arbitrage` as the single composite operation.
5. If the request mentions `watch-*`, explain that streaming is deferred in `v1` and do not invent a workaround.

## Boundaries

- Do not add new analytics, strategy, or automation behavior.
- Do not create extra tools for aliases.
- Do not treat streaming as part of `v1`.
- Do not split `fund-arb` into hand-built order legs in the skill layer.
