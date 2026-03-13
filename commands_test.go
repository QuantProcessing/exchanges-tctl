package main

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestRequireArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		min     int
		wantErr bool
	}{
		{"enough args", []string{"BTC", "0.1"}, 2, false},
		{"more than enough", []string{"BTC", "0.1", "--price", "50000"}, 2, false},
		{"exact", []string{"BTC"}, 1, false},
		{"not enough", []string{}, 1, true},
		{"zero min always passes", []string{}, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireArgs(tt.args, tt.min, "test usage")
			if (err != nil) != tt.wantErr {
				t.Errorf("requireArgs(%v, %d) error = %v, wantErr %v", tt.args, tt.min, err, tt.wantErr)
			}
		})
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantPrice string
		wantArgs  int // remaining args count
	}{
		{
			name:      "with price flag",
			args:      []string{"--price", "50000"},
			wantPrice: "50000",
			wantArgs:  0,
		},
		{
			name:      "price in middle",
			args:      []string{"--reduce-only", "--price", "51000.5", "--other"},
			wantPrice: "51000.5",
			wantArgs:  2,
		},
		{
			name:      "no price flag",
			args:      []string{"--reduce-only"},
			wantPrice: "0",
			wantArgs:  1,
		},
		{
			name:      "empty args",
			args:      []string{},
			wantPrice: "0",
			wantArgs:  0,
		},
		{
			name:      "price flag without value",
			args:      []string{"--price"},
			wantPrice: "0",
			wantArgs:  1,
		},
		{
			name:      "price flag with invalid value",
			args:      []string{"--price", "abc"},
			wantPrice: "0",
			wantArgs:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, remaining := parsePrice(tt.args)
			if price.String() != tt.wantPrice {
				t.Errorf("parsePrice(%v) price = %s, want %s", tt.args, price.String(), tt.wantPrice)
			}
			if len(remaining) != tt.wantArgs {
				t.Errorf("parsePrice(%v) remaining count = %d, want %d", tt.args, len(remaining), tt.wantArgs)
			}
		})
	}
}

func TestParseQty(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantQty  string
		wantArgs int
	}{
		{
			name:     "with qty flag",
			args:     []string{"--qty", "0.5"},
			wantQty:  "0.5",
			wantArgs: 0,
		},
		{
			name:     "no qty flag",
			args:     []string{"--price", "50000"},
			wantQty:  "0",
			wantArgs: 2,
		},
		{
			name:     "empty",
			args:     []string{},
			wantQty:  "0",
			wantArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qty, remaining := parseQty(tt.args)
			if qty.String() != tt.wantQty {
				t.Errorf("parseQty(%v) = %s, want %s", tt.args, qty.String(), tt.wantQty)
			}
			if len(remaining) != tt.wantArgs {
				t.Errorf("parseQty(%v) remaining = %d, want %d", tt.args, len(remaining), tt.wantArgs)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flag     string
		wantBool bool
		wantArgs int
	}{
		{"flag present", []string{"--reduce-only"}, "--reduce-only", true, 0},
		{"flag absent", []string{"--other"}, "--reduce-only", false, 1},
		{"flag in middle", []string{"a", "--reduce-only", "b"}, "--reduce-only", true, 2},
		{"empty args", []string{}, "--reduce-only", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, remaining := parseBool(tt.args, tt.flag)
			if got != tt.wantBool {
				t.Errorf("parseBool(%v, %s) = %v, want %v", tt.args, tt.flag, got, tt.wantBool)
			}
			if len(remaining) != tt.wantArgs {
				t.Errorf("parseBool(%v, %s) remaining = %d, want %d", tt.args, tt.flag, len(remaining), tt.wantArgs)
			}
		})
	}
}

func TestParseDecimal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"integer", "100", "100", false},
		{"decimal", "0.001", "0.001", false},
		{"negative", "-5.5", "-5.5", false},
		{"invalid", "abc", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDecimal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDecimal(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("parseDecimal(%q) = %s, want %s", tt.input, got.String(), tt.want)
			}
		})
	}
}

func TestPriceStr(t *testing.T) {
	tests := []struct {
		name  string
		price decimal.Decimal
		want  string
	}{
		{"zero is dash", decimal.Zero, "-"},
		{"positive", decimal.NewFromInt(50000), "50000"},
		{"decimal", decimal.RequireFromString("95123.45"), "95123.45"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := priceStr(tt.price)
			if got != tt.want {
				t.Errorf("priceStr(%v) = %s, want %s", tt.price, got, tt.want)
			}
		})
	}
}

func TestDecStr(t *testing.T) {
	d := decimal.RequireFromString("123.456")
	if got := decStr(d); got != "123.456" {
		t.Errorf("decStr() = %s, want 123.456", got)
	}
}

func TestParseStringFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flag     string
		wantVal  string
		wantRest int
	}{
		{"found", []string{"--from", "SPOT"}, "--from", "SPOT", 0},
		{"not found", []string{"--other", "val"}, "--from", "", 2},
		{"empty args", nil, "--from", "", 0},
		{"flag without value", []string{"--from"}, "--from", "", 1},
		{"in middle", []string{"a", "--from", "PERP", "b"}, "--from", "PERP", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, rest := parseStringFlag(tt.args, tt.flag)
			if val != tt.wantVal {
				t.Errorf("parseStringFlag() val = %q, want %q", val, tt.wantVal)
			}
			if len(rest) != tt.wantRest {
				t.Errorf("parseStringFlag() remaining = %d items, want %d", len(rest), tt.wantRest)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"10", 10, false},
		{"0", 0, false},
		{"-1", -1, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := parseInt(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
