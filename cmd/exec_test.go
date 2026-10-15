package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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

func TestExecArgsKeepCommandFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want []string
	}{
		{"without separator", []string{"-e", "dev", "exec", "java", "-jar", "app.jar"}, []string{"java", "-jar", "app.jar"}},
		{"with separator", []string{"-e", "dev", "exec", "--", "java", "-jar", "app.jar"}, []string{"java", "-jar", "app.jar"}},
		{"cloudwrap flags after exec", []string{"exec", "-s", "svc", "run.sh", "--port", "8080"}, []string{"run.sh", "--port", "8080"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := newRootCmd()
			var got []string
			for _, c := range root.Commands() {
				if c.Name() == "exec" {
					c.RunE = func(_ *cobra.Command, args []string) error {
						got = args
						return nil
					}
				}
			}
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if strings.Join(got, " ") != strings.Join(tc.want, " ") {
				t.Fatalf("args = %q, want %q", got, tc.want)
			}
		})
	}
}
