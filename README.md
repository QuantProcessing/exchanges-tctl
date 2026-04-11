# tctl — Trader Exchange Control Tool

[中文文档](README_CN.md)

A command-line tool for managing cryptocurrency exchanges. Supports Perp & Spot markets, real-time streaming, and an interactive REPL.

## Install

### Option 1: go install (recommended)

```bash
go install github.com/QuantProcessing/exchanges-tctl@latest
```

### Option 2: Build from source

```bash
go build -o tctl .
```

## Quick Start

```bash
# Get ticker price
tctl ticker BTC

# Limit buy (perp)
tctl buy BTC 0.001 --price 50000

# Market sell
tctl sell ETH 0.1

# Spot mode
tctl -m spot -e BINANCE spot-balances

# WebSocket live ticker
tctl -ws watch-ticker ETH

# Positions as JSON
tctl positions --json
```

## Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-e` | Exchange name | Auto-detect |
| `-m` | Market type: `perp \| spot` | `perp` |
| `-json` | JSON output (AI agent friendly) | `false` |
| `-ws` | Compatibility flag for watch/WebSocket-centric workflows | `false` |
| `-version` | Show version | - |

## Commands

### Market Data

| Command | Alias | Description |
|---------|-------|-------------|
| `ticker <symbol>` | `t` | Get ticker price |
| `orderbook <symbol> [depth]` | `ob` | Order book |
| `trades <symbol> [limit]` | - | Recent trades |
| `klines <symbol> <interval> [limit]` | `kl` | Candlestick data |
| `details <symbol>` | - | Symbol details |
| `fee <symbol>` | - | Fee rate |
| `funding <symbol>` | - | Funding rate (perp) |

### Trading

| Command | Description |
|---------|-------------|
| `buy <symbol> <qty> [flags]` | Place buy order |
| `sell <symbol> <qty> [flags]` | Place sell order |
| `modify <orderID> <symbol> [--price P] [--qty Q]` | Modify order (perp) |
| `cancel <orderID> <symbol>` | Cancel order |
| `cancel-all <symbol>` | Cancel all orders |
| `order <orderID> <symbol>` | Fetch order details |

**Order flags:**

| Flag | Description |
|------|-------------|
| `--price P` | Limit price (market order if omitted) |
| `--tif GTC\|IOC\|FOK\|PO` | Time-in-force |
| `--post-only` | Post-only mode (same as --tif PO) |
| `--reduce-only` | Reduce only |
| `--client-id ID` | Custom order ID |

### Arbitrage

| Command | Alias | Description |
|---------|-------|-------------|
| `fund-arb <symbol> <qty> [flags]` | `fa` | Funding rate arb (buy spot + short perp) |

**Fund-arb flags:**

| Flag | Description |
|------|-------------|
| `--leverage N` | Perp leverage (default: 1) |
| `--close` | Close mode (sell spot + buy perp) |
| `--spot-price P` | Spot limit price (market if omitted) |
| `--perp-price P` | Perp limit price (market if omitted) |
| `--spot-exchange EX` | Spot leg exchange (must be paired with `--perp-exchange`) |
| `--perp-exchange EX` | Perp leg exchange (must be paired with `--spot-exchange`) |

Examples:

```bash
# Same-exchange arb (legacy behavior)
tctl -e BINANCE fund-arb BTC 0.01 --leverage 5

# Cross-exchange arb
tctl fund-arb BTC 0.01 --spot-exchange BINANCE --perp-exchange OKX
```

### Account

| Command | Alias | Description |
|---------|-------|-------------|
| `positions` | `p` | List positions (perp) |
| `orders [symbol]` | `o` | List open orders |
| `balance` | `b` | Show balance |
| `account` | `acc` | Full account info |
| `leverage <symbol> <value>` | `lev` | Set leverage (perp) |
| `spot-balances` | `sb` | Spot balances (spot) |
| `transfer <asset> <amount> [flags]` | - | Asset transfer (spot) |

### Streaming (WebSocket)

| Command | Alias | Description |
|---------|-------|-------------|
| `watch-ticker <symbol>` | `wt` | Live ticker |
| `watch-ob <symbol> [depth]` | `wob` | Live order book |
| `watch-orders` | `wo` | Live order updates |
| `watch-trades <symbol>` | `wtr` | Live trade stream |

## Interactive Mode

Run without a command to enter REPL mode with dynamic exchange/market/session-mode switching:

```bash
$ tctl -e BINANCE
Connected to BINANCE (perp). Type help for commands, exit to quit.
BINANCE/perp(rest)> ticker BTC
BINANCE/perp(rest)> market spot       # switch to spot
✓ Switched to spot market
BINANCE/spot(rest)> spot-balances
BINANCE/spot(rest)> use OKX           # switch exchange
✓ Switched to OKX
OKX/spot(rest)> mode ws               # switch to WebSocket
✓ Switched to ws mode
OKX/spot(ws)> watch-ticker ETH
OKX/spot(ws)> status                  # show session info
OKX/spot(ws)> exit
```

### Session Commands

| Command | Description |
|---------|-------------|
| `use <exchange>` | Switch exchange |
| `market perp\|spot` | Switch market type |
| `mode rest\|ws` | Switch session mode |
| `status` | Show session info |
| `help` | Show help |
| `exit` | Quit |

## AI Agent Integration

All commands support `--json` for structured JSON output:

```bash
tctl positions --json
# → [{"symbol":"BTC","side":"LONG","quantity":"0.5",...}]

tctl -m spot spot-balances --json
# → [{"asset":"BTC","free":"1.5","locked":"0.3","total":"1.8"}]
```

- Exit code: `0` = success, `1` = error
- Error JSON: `{"error":"message"}`
- stdout = data, stderr = diagnostics

## MCP Plugin

This repository now includes a repo-local Codex plugin scaffold under [`plugins/tctl`](plugins/tctl) and a stdio MCP server entrypoint:

```bash
tctl mcp serve
```

Notes:

- `v1` only exposes request/response tools based on existing `tctl` commands.
- Read and mutating tools are both included.
- `fund-arb` is exposed as a standalone composite tool.
- `watch-*` / streaming workflows are intentionally deferred in `v1`.

The repo-local plugin config currently detects whether it was launched from the repo root or from `plugins/tctl`, then runs:

```bash
go run . mcp serve
```

from [`plugins/tctl/.mcp.json`](plugins/tctl/.mcp.json). This removes the previously hardcoded absolute repo path while still letting the plugin work directly from the checkout without requiring a preinstalled `tctl` binary.

## Credential Configuration

`tctl` reads exchange credentials from environment variables. Variable names use the exchange name directly, without the old `EXCHANGES_` prefix.

Examples:

```bash
BINANCE_API_KEY=...
BINANCE_SECRET_KEY=...

HYPERLIQUID_PRIVATE_KEY=...
HYPERLIQUID_ACCOUNT_ADDR=...

EDGEX_PRIVATE_KEY=...
EDGEX_ACCOUNT_ID=...
```

You can configure credentials in either of these ways:

1. Repo-local `.env`

```bash
cp .env.example .env
# edit .env and fill in only the exchanges you use
```

2. `codex mcp add` with `--env`

Binance:

```bash
codex mcp add tctl-binance \
  --env BINANCE_API_KEY=your_binance_api_key \
  --env BINANCE_SECRET_KEY=your_binance_secret_key \
  -- tctl mcp serve
```

Hyperliquid:

```bash
codex mcp add tctl-hyperliquid \
  --env HYPERLIQUID_PRIVATE_KEY=your_hyperliquid_private_key \
  --env HYPERLIQUID_ACCOUNT_ADDR=your_hyperliquid_account_addr \
  -- tctl mcp serve
```

EdgeX:

```bash
codex mcp add tctl-edgex \
  --env EDGEX_PRIVATE_KEY=your_edgex_private_key \
  --env EDGEX_ACCOUNT_ID=your_edgex_account_id \
  -- tctl mcp serve
```

Notes:

- `codex mcp add` only registers a launch command; it does not manage your secrets for you.
- Using `--env KEY=value` can expose secrets in shell history. For local development, `.env` is usually safer and simpler.
- If you configure more than one exchange, pass `-e <EXCHANGE>` in the tool call so `tctl` does not need to auto-detect.

## Development

```bash
# Run tests
go test ./... -v

# Build
go build -o tctl .
```

## Configuration

Configure exchange credentials via `.env` file or environment variables:

```bash
BINANCE_API_KEY=xxx
BINANCE_SECRET_KEY=xxx
OKX_API_KEY=xxx
# Optional: specify quote currency (exchange defaults vary; see .env.example)
# BINANCE_QUOTE_CURRENCY=USDC
```

With `github.com/QuantProcessing/exchanges` `v0.2.13+`, order transport is no longer selected through a shared `OrderMode` toggle. `tctl` still accepts `-ws` and `mode ws` for compatibility with watch-oriented workflows, but standard order commands (`buy`, `sell`, `cancel`, `modify`) follow the adapter's primary non-WS write path.

Supported exchanges: Binance, OKX, Aster, Nado, Lighter, Hyperliquid, StandX, EdgeX, GRVT.

See `.env.example` for full configuration reference.

## License

[MIT](LICENSE)
