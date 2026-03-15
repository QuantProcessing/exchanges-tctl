# tctl — Trader Exchange Control Tool

[中文文档](README_CN.md)

A command-line tool for managing cryptocurrency exchanges. Supports Perp & Spot markets, REST/WebSocket dual-mode transport, and an interactive REPL.

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
| `-ws` | Use WebSocket transport | `false` (REST) |
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

Run without a command to enter REPL mode with dynamic exchange/market/transport switching:

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
| `mode rest\|ws` | Switch transport mode |
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
EXCHANGES_BINANCE_API_KEY=xxx
EXCHANGES_BINANCE_SECRET_KEY=xxx
EXCHANGES_OKX_API_KEY=xxx
# Optional: specify quote currency (default: CEX=USDT, DEX=USDC)
# EXCHANGES_BINANCE_QUOTE_CURRENCY=USDC
```

Supported exchanges: Binance, OKX, Aster, Nado, Lighter, Hyperliquid, StandX, EdgeX, GRVT.

See `.env.example` for full configuration reference.

## License

[MIT](LICENSE)
