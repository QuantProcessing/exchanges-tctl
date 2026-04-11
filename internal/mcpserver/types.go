package mcpserver

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

type ToolGroup string

const (
	ToolGroupRead      ToolGroup = "read"
	ToolGroupMutating  ToolGroup = "mutating"
	ToolGroupStreaming ToolGroup = "streaming"
)

type ToolDefinition struct {
	Name           string
	Command        string
	Group          ToolGroup
	Deferred       bool
	DeferredReason string
	Aliases        []string
}

type ToolCallRequest struct {
	ToolName  string
	Exchange  string
	Market    string
	Arguments map[string]any
}

type CLIInvocation struct {
	Command string
	Args    []string
}

type CLIProcessResult struct {
	ExitCode  int
	Stdout    []byte
	Stderr    []byte
	StartErr  error
	TimedOut  bool
	Cancelled bool
	Duration  time.Duration
}

type CLIResultMetadata struct {
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	Duration string `json:"duration,omitempty"`
}

type DecodedCLIResult struct {
	Success  bool
	Content  json.RawMessage
	Metadata CLIResultMetadata
}

type CLIError struct {
	Code           string
	Message        string
	StructuredBody json.RawMessage
	Metadata       CLIResultMetadata
}

func (e *CLIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func currentTCTLBinary() string {
	if bin := strings.TrimSpace(os.Getenv("TCTL_MCP_TCTL_BIN")); bin != "" {
		return bin
	}
	exe, err := os.Executable()
	if err == nil && exe != "" {
		return exe
	}
	return "tctl"
}

func normalizeJSON(raw []byte) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	return json.RawMessage(bytesTrimSpace(raw))
}

func bytesTrimSpace(in []byte) []byte {
	return []byte(strings.TrimSpace(string(in)))
}

func trimString(in []byte) string {
	return strings.TrimSpace(string(in))
}

func newCLIError(code, message string, body json.RawMessage, metadata CLIResultMetadata) error {
	return &CLIError{
		Code:           code,
		Message:        message,
		StructuredBody: body,
		Metadata:       metadata,
	}
}

func metadataFromResult(result CLIProcessResult) CLIResultMetadata {
	md := CLIResultMetadata{}
	if stdout := trimString(result.Stdout); stdout != "" {
		md.Stdout = stdout
	}
	if stderr := trimString(result.Stderr); stderr != "" {
		md.Stderr = stderr
	}
	if result.Duration > 0 {
		md.Duration = result.Duration.String()
	}
	return md
}
