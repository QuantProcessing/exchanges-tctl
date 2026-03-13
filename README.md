# tctl — Trader Exchange Control Tool

交易所命令行控制工具，支持 Perp 和 Spot 市场、REST/WebSocket 双模式、交互式 REPL。

## 安装

### 方式一：go install（推荐）

```bash
go install github.com/QuantProcessing/exchanges-tctl@latest
```

### 方式二：本地编译

```bash
go build -o tctl .
```

## 快速开始

```bash
# 查看价格
tctl ticker BTC

# 限价买入（perp）
tctl buy BTC 0.001 --price 50000

# 市价卖出
tctl sell ETH 0.1

# Spot 模式
tctl -m spot -e BINANCE spot-balances

# WebSocket 实时行情
tctl -ws watch-ticker ETH

# 查看仓位（JSON 输出）
tctl positions --json
```

## 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-e` | 交易所名称 | 自动检测 |
| `-m` | 市场类型: `perp \| spot` | `perp` |
| `-json` | JSON 输出（AI agent 友好） | `false` |
| `-ws` | 使用 WebSocket | `false`（REST） |
| `-version` | 显示版本 | - |

## 命令列表

### 行情

| 命令 | 别名 | 说明 |
|------|------|------|
| `ticker <symbol>` | `t` | 获取价格 |
| `orderbook <symbol> [depth]` | `ob` | 订单簿 |
| `trades <symbol> [limit]` | - | 最近成交 |
| `klines <symbol> <interval> [limit]` | `kl` | K线数据 |
| `details <symbol>` | - | 交易对详情 |
| `fee <symbol>` | - | 手续费率 |
| `funding <symbol>` | - | 资金费率 (perp) |

### 交易

| 命令 | 说明 |
|------|------|
| `buy <symbol> <qty> [flags]` | 买入 |
| `sell <symbol> <qty> [flags]` | 卖出 |
| `modify <orderID> <symbol> [--price P] [--qty Q]` | 修改订单 (perp) |
| `cancel <orderID> <symbol>` | 撤销订单 |
| `cancel-all <symbol>` | 撤销全部 |
| `order <orderID> <symbol>` | 查询单个订单 |

**交易标志:**

| 标志 | 说明 |
|------|------|
| `--price P` | 限价（不指定则市价） |
| `--tif GTC\|IOC\|FOK\|PO` | 时间条件 |
| `--post-only` | 挂单模式（等同 --tif PO） |
| `--reduce-only` | 仅减仓 |
| `--client-id ID` | 自定义订单 ID |

### 账户

| 命令 | 别名 | 说明 |
|------|------|------|
| `positions` | `p` | 持仓列表 (perp) |
| `orders [symbol]` | `o` | 挂单列表 |
| `balance` | `b` | 余额 |
| `account` | `acc` | 完整账户 |
| `leverage <symbol> <value>` | `lev` | 设置杠杆 (perp) |
| `spot-balances` | `sb` | 现货余额 (spot) |
| `transfer <asset> <amount> [flags]` | - | 划转资产 (spot) |

### 流式订阅 (WSS)

| 命令 | 别名 | 说明 |
|------|------|------|
| `watch-ticker <symbol>` | `wt` | 实时报价 |
| `watch-ob <symbol> [depth]` | `wob` | 实时订单簿 |
| `watch-orders` | `wo` | 实时订单更新 |
| `watch-trades <symbol>` | `wtr` | 实时成交 |

## 交互模式

不带命令运行进入 REPL，支持动态切换交易所/市场/传输模式：

```bash
$ tctl -e BINANCE
Connected to BINANCE (perp). Type help for commands, exit to quit.
BINANCE/perp(rest)> ticker BTC
BINANCE/perp(rest)> market spot       # 切换到现货
✓ Switched to spot market
BINANCE/spot(rest)> spot-balances
BINANCE/spot(rest)> use OKX           # 切换交易所
✓ Switched to OKX
OKX/spot(rest)> mode ws               # 切换到 WebSocket
✓ Switched to ws mode
OKX/spot(ws)> watch-ticker ETH
OKX/spot(ws)> status                  # 查看当前会话状态
OKX/spot(ws)> exit
```

### 交互模式特有命令

| 命令 | 说明 |
|------|------|
| `use <exchange>` | 切换交易所 |
| `market perp\|spot` | 切换市场类型 |
| `mode rest\|ws` | 切换传输模式 |
| `status` | 显示当前会话信息 |
| `help` | 帮助 |
| `exit` | 退出 |

## AI Agent 集成

所有命令支持 `--json` 输出结构化 JSON：

```bash
tctl positions --json
# → [{"symbol":"BTC","side":"LONG","quantity":"0.5",...}]

tctl -m spot spot-balances --json
# → [{"asset":"BTC","free":"1.5","locked":"0.3","total":"1.8"}]
```

- Exit code: `0` = 成功，`1` = 错误
- Error as JSON: `{"error":"message"}`
- stdout = 数据，stderr = 诊断

## 开发

```bash
# 运行测试
go test ./... -v

# 本地编译
go build -o tctl .
```

## 配置

通过 `.env` 文件或环境变量配置交易所凭证：

```bash
EXCHANGES_BINANCE_API_KEY=xxx
EXCHANGES_BINANCE_SECRET_KEY=xxx
EXCHANGES_OKX_API_KEY=xxx
# 可选：指定报价币种（默认 CEX=USDT, DEX=USDC）
# EXCHANGES_BINANCE_QUOTE_CURRENCY=USDC
```

支持的交易所：Binance, OKX, Aster, Nado, Lighter, Hyperliquid, StandX, EdgeX, GRVT。

详见 `.env.example`。
