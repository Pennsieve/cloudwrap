package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/pennsieve/cloudwrap/internal/params"
)

func TestRunCommandForwardsExitCode(t *testing.T) {
	err := runCommand(context.Background(), []string{"sh", "-c", "exit 7"}, nil)
	var ee *execError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *execError, got %v", err)
	}
	if ee.code != 7 {
		t.Fatalf("expected exit code 7, got %d", ee.code)
	}
	if got := exitCodeFor(err); got != 7 {
		t.Fatalf("exitCodeFor = %d, want 7", got)
	}
}

func TestRunCommandSuccess(t *testing.T) {
	if err := runCommand(context.Background(), []string{"true"}, nil); err != nil {
		t.Fatalf("expected nil error for `true`, got %v", err)
	}
}

func TestRunCommandInjectsEnv(t *testing.T) {
	// `sh -c 'test "$FOO" = bar'` exits 0 only if FOO was injected.
	pairs := []params.Pair{{Key: "FOO", Value: "bar"}}
	if err := runCommand(context.Background(), []string{"sh", "-c", `test "$FOO" = bar`}, pairs); err != nil {
		t.Fatalf("expected injected env FOO=bar, got %v", err)
	}
}

func TestExitCodeForGenericError(t *testing.T) {
	if got := exitCodeFor(errors.New("boom")); got != 1 {
		t.Fatalf("exitCodeFor(generic) = %d, want 1", got)
	}
}
