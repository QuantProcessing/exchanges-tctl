package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestOutputJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	outputJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, `"key": "value"`) {
		t.Errorf("outputJSON() output = %q, want to contain key:value", output)
	}
}

func TestOutputTable(t *testing.T) {
	output := captureOutput(func() {
		outputTable(
			[]string{"Name", "Price"},
			[][]string{
				{"BTC", "50000"},
				{"ETH", "3000"},
			},
		)
	})

	if !strings.Contains(output, "Name") {
		t.Error("outputTable() missing header")
	}
	if !strings.Contains(output, "BTC") {
		t.Error("outputTable() missing BTC row")
	}
	if !strings.Contains(output, "ETH") {
		t.Error("outputTable() missing ETH row")
	}
}

func TestDecStrFormatting(t *testing.T) {
	tests := []struct {
		input decimal.Decimal
		want  string
	}{
		{decimal.NewFromInt(0), "0"},
		{decimal.NewFromFloat(1.5), "1.5"},
		{decimal.RequireFromString("123456789.123456789"), "123456789.123456789"},
	}

	for _, tt := range tests {
		got := decStr(tt.input)
		if got != tt.want {
			t.Errorf("decStr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ============================================================================
// Color function tests
// ============================================================================

func TestColorSide(t *testing.T) {
	tests := []struct {
		input string
		color string // expected ANSI prefix
	}{
		{"BUY", ansiGreen},
		{"SELL", ansiRed},
		{"LONG", ansiGreen},
		{"SHORT", ansiRed},
		{"UNKNOWN", ""}, // no color
	}

	for _, tt := range tests {
		result := colorSide(tt.input)
		if tt.color != "" && !strings.HasPrefix(result, tt.color) {
			t.Errorf("colorSide(%q) should start with color code", tt.input)
		}
		if tt.color == "" && result != tt.input {
			t.Errorf("colorSide(%q) = %q, want %q", tt.input, result, tt.input)
		}
	}
}

func TestColorPnL(t *testing.T) {
	positive := colorPnL(decimal.RequireFromString("100"))
	if !strings.Contains(positive, "+100") {
		t.Error("positive PnL should have + prefix")
	}
	if !strings.HasPrefix(positive, ansiGreen) {
		t.Error("positive PnL should be green")
	}

	negative := colorPnL(decimal.RequireFromString("-50"))
	if !strings.Contains(negative, "-50") {
		t.Error("negative PnL should show negative value")
	}
	if !strings.HasPrefix(negative, ansiRed) {
		t.Error("negative PnL should be red")
	}

	zero := colorPnL(decimal.Zero)
	if zero != "0" {
		t.Errorf("zero PnL = %q, want '0'", zero)
	}
}

func TestColorStatus(t *testing.T) {
	tests := []struct {
		status string
		color  string
	}{
		{"FILLED", ansiGreen},
		{"CANCELLED", ansiRed},
		{"REJECTED", ansiRed},
		{"PARTIALLY_FILLED", ansiYellow},
		{"NEW", ansiCyan},
		{"PENDING", ansiCyan},
		{"UNKNOWN", ""},
	}

	for _, tt := range tests {
		result := colorStatus(tt.status)
		if tt.color != "" && !strings.HasPrefix(result, tt.color) {
			t.Errorf("colorStatus(%q) should be colored", tt.status)
		}
	}
}

func TestOutputSuccess(t *testing.T) {
	output := captureOutput(func() {
		outputSuccess("test %s %d", "abc", 123)
	})
	if !strings.Contains(output, "abc") {
		t.Error("outputSuccess should contain formatted text")
	}
	if !strings.Contains(output, "✓") {
		t.Error("outputSuccess should contain check mark")
	}
}

func TestOutputInfo(t *testing.T) {
	output := captureOutput(func() {
		outputInfo("info message")
	})
	if !strings.Contains(output, "info message") {
		t.Error("outputInfo should contain the message")
	}
	if !strings.Contains(output, "ℹ") {
		t.Error("outputInfo should contain info icon")
	}
}

func TestOutputError(t *testing.T) {
	// outputError writes to stderr, but we test it doesn't panic
	outputError("test error %d", 42)
}

func TestColorDim(t *testing.T) {
	result := colorDim("dimmed")
	if !strings.Contains(result, "dimmed") {
		t.Error("colorDim should contain the text")
	}
	if !strings.HasPrefix(result, "\033[2m") {
		t.Error("colorDim should apply dim ANSI code")
	}
}

func TestOutputTableEmpty(t *testing.T) {
	output := captureOutput(func() {
		outputTable([]string{"A", "B"}, nil)
	})
	if !strings.Contains(output, "A") {
		t.Error("outputTable with nil rows should still print headers")
	}
}

func TestColorSideEdgeCases(t *testing.T) {
	// Test lowercase passing through
	result := colorSide("other")
	if result != "other" {
		t.Errorf("colorSide('other') = %q, want 'other'", result)
	}
}

func TestColorPnLZero(t *testing.T) {
	result := colorPnL(decimal.Zero)
	if result != "0" {
		t.Errorf("colorPnL(0) = %q, want '0'", result)
	}
}

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain", "hello", 5},
		{"colored", colorGreen("BUY"), 3},
		{"bold", colorBold("TEXT"), 4},
		{"empty", "", 0},
		{"multi color", colorGreen("A") + " " + colorRed("B"), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleLen(tt.input)
			if got != tt.want {
				t.Errorf("visibleLen(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestValOrDash(t *testing.T) {
	if got := valOrDash(""); got != "-" {
		t.Errorf("valOrDash('') = %q, want '-'", got)
	}
	if got := valOrDash("GTC"); got != "GTC" {
		t.Errorf("valOrDash('GTC') = %q, want 'GTC'", got)
	}
}

func TestOutputTableAlignment(t *testing.T) {
	output := captureOutput(func() {
		outputTable(
			[]string{"Symbol", "Side", "Price"},
			[][]string{
				{"BTC", colorGreen("BUY"), "95000"},
				{"ETH", colorRed("SELL"), "3000"},
			},
		)
	})

	// Both rows should have consistent formatting
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), output)
	}
	// The key test: BUY and SELL should visually align despite different ANSI lengths
	// We can't easily test pixel alignment, but we can verify structure
	if !strings.Contains(lines[1], "BTC") {
		t.Error("row 1 should contain BTC")
	}
	if !strings.Contains(lines[2], "ETH") {
		t.Error("row 2 should contain ETH")
	}
}
