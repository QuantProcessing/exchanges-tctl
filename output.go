package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

// ============================================================================
// ANSI Color Helpers
// ============================================================================

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
)

func colorRed(s string) string    { return ansiRed + s + ansiReset }
func colorGreen(s string) string  { return ansiGreen + s + ansiReset }
func colorYellow(s string) string { return ansiYellow + s + ansiReset }
func colorCyan(s string) string   { return ansiCyan + s + ansiReset }
func colorBold(s string) string   { return ansiBold + s + ansiReset }
func colorDim(s string) string    { return ansiDim + s + ansiReset }

// colorSide returns green for BUY/LONG, red for SELL/SHORT.
func colorSide(side string) string {
	switch strings.ToUpper(side) {
	case "BUY", "LONG":
		return colorGreen(side)
	case "SELL", "SHORT":
		return colorRed(side)
	default:
		return side
	}
}

// colorPnL returns green for positive, red for negative PnL.
func colorPnL(d decimal.Decimal) string {
	s := d.String()
	if d.IsPositive() {
		return colorGreen("+" + s)
	}
	if d.IsNegative() {
		return colorRed(s)
	}
	return s
}

// colorStatus returns colored order status.
func colorStatus(status string) string {
	switch status {
	case "FILLED":
		return colorGreen(status)
	case "CANCELLED", "REJECTED":
		return colorRed(status)
	case "PARTIALLY_FILLED":
		return colorYellow(status)
	case "NEW", "PENDING":
		return colorCyan(status)
	default:
		return status
	}
}

// ============================================================================
// Output Functions
// ============================================================================

func outputJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// ansiRegex matches ANSI escape sequences for stripping.
var ansiRegex = regexp.MustCompile(`\033\[[0-9;]*m`)

// visibleLen returns the visible character count, excluding ANSI escape codes.
func visibleLen(s string) int {
	clean := ansiRegex.ReplaceAllString(s, "")
	return utf8.RuneCountInString(clean)
}

// outputTable prints a formatted table with proper ANSI-aware column alignment.
func outputTable(headers []string, rows [][]string) {
	cols := len(headers)
	if cols == 0 {
		return
	}

	// Calculate max visible width for each column
	widths := make([]int, cols)
	for i, h := range headers {
		widths[i] = visibleLen(h)
	}
	for _, row := range rows {
		for i := 0; i < cols && i < len(row); i++ {
			if vl := visibleLen(row[i]); vl > widths[i] {
				widths[i] = vl
			}
		}
	}

	// Print header
	for i, h := range headers {
		pad := widths[i] - visibleLen(h) + 2
		fmt.Print(h + strings.Repeat(" ", pad))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i := 0; i < cols; i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			pad := widths[i] - visibleLen(cell) + 2
			if pad < 2 {
				pad = 2
			}
			fmt.Print(cell + strings.Repeat(" ", pad))
		}
		fmt.Println()
	}
}

func outputSuccess(format string, args ...interface{}) {
	fmt.Printf(colorGreen("✓")+" "+format+"\n", args...)
}

func outputError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, colorRed("✗")+" "+format+"\n", args...)
}

func outputInfo(format string, args ...interface{}) {
	fmt.Printf(colorCyan("ℹ")+" "+format+"\n", args...)
}

func decStr(d decimal.Decimal) string {
	return d.String()
}

// valOrDash returns "-" if the string is empty.
func valOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
