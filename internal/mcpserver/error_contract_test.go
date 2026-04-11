package mcpserver

import (
	"errors"
	"testing"
	"time"
)

func TestDecodeCLIResultErrorContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		result      CLIProcessResult
		wantSuccess bool
		wantCode    string
		wantMessage string
		wantBody    string
	}{
		{
			name: "success parses stdout json and keeps stderr as metadata",
			result: CLIProcessResult{
				ExitCode: 0,
				Stdout:   []byte(`{"symbol":"SOLUSDT","price":"123.45"}`),
				Stderr:   []byte("warning: dry-run metadata"),
			},
			wantSuccess: true,
		},
		{
			name: "fatal json is a structured failure",
			result: CLIProcessResult{
				ExitCode: 1,
				Stdout:   []byte(`{"error":"failed to create BINANCE adapter"}`),
				Stderr:   []byte("adapter init failed"),
			},
			wantCode:    "cli_command_failed",
			wantMessage: "failed to create BINANCE adapter",
			wantBody:    `{"error":"failed to create BINANCE adapter"}`,
		},
		{
			name: "stderr only failure uses stderr as primary message",
			result: CLIProcessResult{
				ExitCode: 2,
				Stderr:   []byte("invalid quantity: abc"),
			},
			wantCode:    "cli_command_failed",
			wantMessage: "invalid quantity: abc",
		},
		{
			name: "invalid success payload is adapter failure",
			result: CLIProcessResult{
				ExitCode: 0,
				Stdout:   []byte("not-json"),
				Stderr:   []byte("debug trace"),
			},
			wantCode:    "invalid_cli_json_output",
			wantMessage: "CLI reported success with invalid JSON output",
		},
		{
			name: "startup failure is distinct from command failure",
			result: CLIProcessResult{
				StartErr: errors.New("exec: \"tctl\": executable file not found in $PATH"),
			},
			wantCode:    "cli_start_failed",
			wantMessage: "failed to start tctl process",
		},
		{
			name: "timeout is distinct from command failure",
			result: CLIProcessResult{
				TimedOut: true,
				Stdout:   []byte(`{"partial":"output"}`),
				Stderr:   []byte("context deadline exceeded"),
				Duration: 5 * time.Second,
			},
			wantCode:    "cli_timeout",
			wantMessage: "tctl process timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := DecodeCLIResult(tt.result)
			if tt.wantSuccess {
				if err != nil {
					t.Fatalf("DecodeCLIResult() unexpected error = %v", err)
				}
				if !got.Success {
					t.Fatalf("DecodeCLIResult() success = false, want true")
				}
				if string(got.Content) != `{"symbol":"SOLUSDT","price":"123.45"}` {
					t.Fatalf("DecodeCLIResult() content = %s", got.Content)
				}
				if got.Metadata.Stderr != "warning: dry-run metadata" {
					t.Fatalf("DecodeCLIResult() stderr metadata = %q, want %q", got.Metadata.Stderr, "warning: dry-run metadata")
				}
				return
			}

			if err == nil {
				t.Fatalf("DecodeCLIResult() error = nil, want failure")
			}

			var cliErr *CLIError
			if !errors.As(err, &cliErr) {
				t.Fatalf("DecodeCLIResult() error type = %T, want *CLIError", err)
			}

			if cliErr.Code != tt.wantCode {
				t.Fatalf("DecodeCLIResult() error code = %q, want %q", cliErr.Code, tt.wantCode)
			}
			if cliErr.Message != tt.wantMessage {
				t.Fatalf("DecodeCLIResult() error message = %q, want %q", cliErr.Message, tt.wantMessage)
			}
			if tt.wantBody != "" && string(cliErr.StructuredBody) != tt.wantBody {
				t.Fatalf("DecodeCLIResult() structured body = %s, want %s", cliErr.StructuredBody, tt.wantBody)
			}
			if tt.wantCode == "invalid_cli_json_output" {
				if cliErr.Metadata.Stdout != "not-json" {
					t.Fatalf("DecodeCLIResult() invalid json stdout metadata = %q", cliErr.Metadata.Stdout)
				}
				if cliErr.Metadata.Stderr != "debug trace" {
					t.Fatalf("DecodeCLIResult() invalid json stderr metadata = %q", cliErr.Metadata.Stderr)
				}
			}
		})
	}
}
