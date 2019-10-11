package logging

import (
	"context"
	"strings"
	"testing"
)

func TestConsoleWithNoAlternateWriter(t *testing.T) {

	cl := new(ConsoleErrorLogger)
	cl.LogDebugf("")

	cl.LogErrorf("")

	cl = new(ConsoleErrorLogger)
	cl.LogErrorfCtx(context.Background(), "")
}

func TestMessagesBelowErrorNotWritten(t *testing.T) {

	cl := new(ConsoleErrorLogger)

	sb := new(strings.Builder)

	cl.w = sb

	ctx := context.Background()

	m := "MESSAGE"

	cl.LogTracefCtx(ctx, m)
	cl.LogTracef(m)
	cl.LogDebugfCtx(ctx, m)
	cl.LogDebugf(m)
	cl.LogWarnfCtx(ctx, m)
	cl.LogWarnf(m)
	cl.LogInfofCtx(ctx, m)
	cl.LogInfof(m)

	cl.LogAtLevelfCtx(ctx, Debug, DebugLabel, m)
	cl.LogAtLevelf(Trace, TraceLabel, m)

	if len(sb.String()) > 0 {
		t.Errorf("Output buffer should be empty but is %s", sb.String())
	}

}

func TestMessagesAboveErrorWritten(t *testing.T) {
	cl := new(ConsoleErrorLogger)

	sb := new(strings.Builder)

	cl.w = sb

	ctx := context.Background()

	m := "MESSAGE"
	mnl := m + "\n"

	cl.LogErrorf(m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogErrorfCtx(ctx, m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogErrorfCtxWithTrace(ctx, m)

	if !strings.HasPrefix(sb.String(), mnl) {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogErrorfWithTrace(m)

	if !strings.HasPrefix(sb.String(), mnl) {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogFatalf(m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogFatalfCtx(ctx, m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogAtLevelfCtx(ctx, Fatal, FatalLabel, m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

	sb.Reset()

	cl.LogAtLevelf(Fatal, FatalLabel, m)

	if sb.String() != mnl {
		t.Errorf("Expected buffer to contain %s but contained %s", m, sb.String())
	}

}
