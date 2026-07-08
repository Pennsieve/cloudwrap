package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/pennsieve/cloudwrap/internal/params"
)

// execError carries the exit code of a failed wrapped command so Execute can
// forward it as cloudwrap's own exit code.
type execError struct {
	code int
}

func (e *execError) Error() string {
	return "wrapped command exited with a non-zero status"
}

// exitCodeFor maps an error returned from Execute to a process exit code. A
// wrapped-command failure forwards that command's code; anything else is a
// generic failure (1).
func exitCodeFor(err error) int {
	var ee *execError
	if errors.As(err, &ee) {
		return ee.code
	}
	return 1
}

func newExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec -- <command> [args...]",
		Short: "Execute a command with the configuration injected as environment variables.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfgs, err := configs()
			if err != nil {
				return err
			}
			client, err := newClient(ctx)
			if err != nil {
				return err
			}
			pairs, err := fetchMerged(ctx, client, cfgs)
			if err != nil {
				return err
			}

			return runCommand(ctx, args, pairs)
		},
	}
}

// runCommand spawns args[0] with args[1:], augmenting the current environment
// with the fetched pairs, forwarding signals to the child, and forwarding its
// exit code.
func runCommand(ctx context.Context, args []string, pairs []params.Pair) error {
	child := exec.CommandContext(ctx, args[0], args[1:]...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	env := os.Environ()
	for _, p := range pairs {
		env = append(env, p.Key+"="+p.Value)
	}
	child.Env = env

	if err := child.Start(); err != nil {
		return err
	}

	// Forward termination signals to the child instead of dying ourselves,
	// so the wrapped command controls its own shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)
	go func() {
		for s := range sigs {
			if child.Process != nil {
				_ = child.Process.Signal(s)
			}
		}
	}()

	err := child.Wait()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		if code < 0 {
			code = 1 // terminated by signal; report generic failure
		}
		return &execError{code: code}
	}
	return err
}
