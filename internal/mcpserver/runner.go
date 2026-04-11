package mcpserver

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

func ExecuteTool(ctx context.Context, req ToolCallRequest) (*DecodedCLIResult, error) {
	invocation, err := BuildCLIInvocation(req)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, invocation.Command, invocation.Args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	started := time.Now()
	runErr := cmd.Run()

	result := CLIProcessResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: time.Since(started),
	}

	switch {
	case ctx.Err() == context.DeadlineExceeded:
		result.TimedOut = true
	case ctx.Err() == context.Canceled:
		result.Cancelled = true
	case runErr == nil:
		result.ExitCode = 0
	case cmd.ProcessState != nil:
		result.ExitCode = cmd.ProcessState.ExitCode()
	default:
		result.StartErr = runErr
	}

	return DecodeCLIResult(result)
}
