package mcpserver

import (
	"encoding/json"
	"fmt"
)

func DecodeCLIResult(result CLIProcessResult) (*DecodedCLIResult, error) {
	metadata := metadataFromResult(result)

	if result.StartErr != nil {
		return nil, newCLIError("cli_start_failed", "failed to start tctl process", nil, metadata)
	}
	if result.Cancelled {
		return nil, newCLIError("cli_cancelled", "tctl process was cancelled", nil, metadata)
	}
	if result.TimedOut {
		return nil, newCLIError("cli_timeout", "tctl process timed out", nil, metadata)
	}

	stdout := normalizeJSON(result.Stdout)
	stderr := trimString(result.Stderr)

	if result.ExitCode == 0 {
		if !json.Valid(stdout) {
			return nil, newCLIError("invalid_cli_json_output", "CLI reported success with invalid JSON output", nil, metadata)
		}
		return &DecodedCLIResult{
			Success:  true,
			Content:  stdout,
			Metadata: metadata,
		}, nil
	}

	if len(stdout) > 0 && json.Valid(stdout) {
		message := extractErrorMessage(stdout)
		if message == "" {
			message = "tctl command failed"
		}
		return nil, newCLIError("cli_command_failed", message, stdout, metadata)
	}

	if stderr != "" {
		return nil, newCLIError("cli_command_failed", stderr, nil, metadata)
	}

	return nil, newCLIError("cli_command_failed", fmt.Sprintf("tctl command failed with exit code %d", result.ExitCode), nil, metadata)
}

func extractErrorMessage(payload json.RawMessage) string {
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return ""
	}
	if msg, ok := body["error"].(string); ok {
		return msg
	}
	return ""
}
