package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	exchanges "github.com/QuantProcessing/exchanges"
	"github.com/QuantProcessing/exchanges/aster"
	"github.com/QuantProcessing/exchanges/binance"
	"github.com/QuantProcessing/exchanges/edgex"
	"github.com/QuantProcessing/exchanges/grvt"
	"github.com/QuantProcessing/exchanges/hyperliquid"
	"github.com/QuantProcessing/exchanges/lighter"
	"github.com/QuantProcessing/exchanges/nado"
	"github.com/QuantProcessing/exchanges/okx"
	"github.com/QuantProcessing/exchanges/standx"
	"go.uber.org/zap"
)

// createAdapter creates an exchange adapter by name using env vars for credentials.
func createAdapter(ctx context.Context, name string, marketType exchanges.MarketType, logger *zap.SugaredLogger) (exchanges.Exchange, error) {
	name = strings.ToUpper(strings.TrimSpace(name))
	prefix := "EXCHANGES_" + name + "_"
	env := func(key string) string { return os.Getenv(prefix + key) }
	quote := exchanges.QuoteCurrency(env("QUOTE_CURRENCY"))

	switch name {
	case "BINANCE":
		if marketType == exchanges.MarketTypeSpot {
			return binance.NewSpotAdapter(ctx, binance.Options{
				APIKey:        env("API_KEY"),
				SecretKey:     env("SECRET_KEY"),
				QuoteCurrency: quote,
				Logger:        logger,
			})
		}
		return binance.NewAdapter(ctx, binance.Options{
			APIKey:        env("API_KEY"),
			SecretKey:     env("SECRET_KEY"),
			QuoteCurrency: quote,
			Logger:        logger,
		})
	case "OKX":
		opts := okx.Options{
			APIKey:        env("API_KEY"),
			SecretKey:     env("SECRET_KEY"),
			Passphrase:    env("PASSPHRASE"),
			QuoteCurrency: quote,
			Logger:        logger,
		}
		if marketType == exchanges.MarketTypeSpot {
			return okx.NewSpotAdapter(ctx, opts)
		}
		return okx.NewAdapter(ctx, opts)
	case "ASTER":
		opts := aster.Options{
			APIKey:        env("API_KEY"),
			SecretKey:     env("SECRET_KEY"),
			QuoteCurrency: quote,
			Logger:        logger,
		}
		if marketType == exchanges.MarketTypeSpot {
			return aster.NewSpotAdapter(ctx, opts)
		}
		return aster.NewAdapter(ctx, opts)
	case "NADO":
		opts := nado.Options{
			PrivateKey:     env("PRIVATE_KEY"),
			SubAccountName: env("SUB_ACCOUNT_NAME"),
			QuoteCurrency:  quote,
			Logger:         logger,
		}
		if marketType == exchanges.MarketTypeSpot {
			return nado.NewSpotAdapter(ctx, opts)
		}
		return nado.NewAdapter(ctx, opts)
	case "LIGHTER":
		opts := lighter.Options{
			PrivateKey:    env("PRIVATE_KEY"),
			AccountIndex:  env("ACCOUNT_INDEX"),
			KeyIndex:      env("KEY_INDEX"),
			RoToken:       env("RO_TOKEN"),
			QuoteCurrency: quote,
			Logger:        logger,
		}
		if marketType == exchanges.MarketTypeSpot {
			return lighter.NewSpotAdapter(ctx, opts)
		}
		return lighter.NewAdapter(ctx, opts)
	case "HYPERLIQUID":
		opts := hyperliquid.Options{
			PrivateKey:    env("PRIVATE_KEY"),
			AccountAddr:   env("ACCOUNT_ADDR"),
			QuoteCurrency: quote,
			Logger:        logger,
		}
		if marketType == exchanges.MarketTypeSpot {
			return hyperliquid.NewSpotAdapter(ctx, opts)
		}
		return hyperliquid.NewAdapter(ctx, opts)
	case "STANDX":
		return standx.NewAdapter(ctx, standx.Options{
			PrivateKey:    env("PRIVATE_KEY"),
			QuoteCurrency: quote,
			Logger:        logger,
		})
	case "EDGEX":
		return edgex.NewAdapter(ctx, edgex.Options{
			PrivateKey:    env("PRIVATE_KEY"),
			AccountID:     env("ACCOUNT_ID"),
			QuoteCurrency: quote,
			Logger:        logger,
		})
	case "GRVT":
		return grvt.NewAdapter(ctx, grvt.Options{
			APIKey:        env("API_KEY"),
			SubAccountID:  env("SUB_ACCOUNT_ID"),
			PrivateKey:    env("PRIVATE_KEY"),
			QuoteCurrency: quote,
			Logger:        logger,
		})
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", name)
	}
}

// createRESTAdapter is kept for CLI compatibility.
// Since exchanges v0.2.13 removed the shared OrderMode toggle, the adapter's
// unsuffixed write methods are already its primary non-WS write path.
func createRESTAdapter(ctx context.Context, name string, marketType exchanges.MarketType, logger *zap.SugaredLogger) (exchanges.Exchange, error) {
	return createAdapter(ctx, name, marketType, logger)
}

// credentialChecks maps exchange names to the env key that indicates the exchange is configured.
var credentialChecks = map[string]string{
	"BINANCE":     "API_KEY",
	"OKX":         "API_KEY",
	"ASTER":       "API_KEY",
	"LIGHTER":     "PRIVATE_KEY",
	"EDGEX":       "PRIVATE_KEY",
	"GRVT":        "API_KEY",
	"NADO":        "PRIVATE_KEY",
	"HYPERLIQUID": "PRIVATE_KEY",
	"STANDX":      "PRIVATE_KEY",
}

// configuredExchanges returns the names of exchanges that have credentials in env.
func configuredExchanges() []string {
	var result []string
	for name, key := range credentialChecks {
		if os.Getenv("EXCHANGES_"+name+"_"+key) != "" {
			result = append(result, name)
		}
	}
	return result
}

// resolveExchange auto-detects the exchange if only one is configured.
func resolveExchange(explicit string) string {
	if explicit != "" {
		return strings.ToUpper(explicit)
	}
	configured := configuredExchanges()
	if len(configured) == 1 {
		return configured[0]
	}
	if len(configured) > 1 {
		fmt.Fprintf(os.Stderr, "Multiple exchanges configured: %v. Use -e to specify.\n", configured)
	}
	return ""
}
