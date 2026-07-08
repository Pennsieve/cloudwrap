// Package cmd implements the cloudwrap command-line interface.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pennsieve/cloudwrap/internal/awsssm"
	"github.com/pennsieve/cloudwrap/internal/config"
	"github.com/pennsieve/cloudwrap/internal/params"
)

var (
	flagEnvironment string
	flagService     string
	version         = "dev"
)

// SetVersion sets the version string reported by `cloudwrap --version`.
func SetVersion(v string) {
	if v != "" {
		version = v
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "cloudwrap",
		Short: "Interfaces with AWS to provide an opinionated way to manage application configuration.",
		Long: "cloudwrap fetches configuration from AWS SSM Parameter Store using resource paths " +
			"of the form /{environment}/{service}/key and exposes it for inspection, printing, or " +
			"injection into a wrapped command as environment variables.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Flags with environment-variable fallbacks, matching the original
	// CLOUDWRAP_ENVIRONMENT / CLOUDWRAP_SERVICE behavior.
	root.PersistentFlags().StringVarP(&flagEnvironment, "environment", "e", os.Getenv("CLOUDWRAP_ENVIRONMENT"),
		"Deployment environment (env: CLOUDWRAP_ENVIRONMENT)")
	root.PersistentFlags().StringVarP(&flagService, "service", "s", os.Getenv("CLOUDWRAP_SERVICE"),
		"Service name; comma-separated for multiple. Left-to-right ordering resolves key conflicts (env: CLOUDWRAP_SERVICE)")

	root.AddCommand(newDescribeCmd(), newStdoutCmd(), newFileCmd(), newExecCmd())
	return root
}

// Execute runs the CLI and returns a process exit code.
func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cloudwrap:", err)
		return exitCodeFor(err)
	}
	return 0
}

// configs validates the required inputs and returns one Config per service, in
// command-line order.
func configs() ([]config.Config, error) {
	if flagEnvironment == "" {
		return nil, fmt.Errorf("environment is required (--environment/-e or CLOUDWRAP_ENVIRONMENT)")
	}
	if flagService == "" {
		return nil, fmt.Errorf("service is required (--service/-s or CLOUDWRAP_SERVICE)")
	}

	var cfgs []config.Config
	for _, s := range strings.Split(flagService, ",") {
		cfgs = append(cfgs, config.New(flagEnvironment, s))
	}
	return cfgs, nil
}

// fetchMerged fetches parameter values for every configured service and merges
// them into ordered env-var pairs (left-most service wins on conflict).
func fetchMerged(ctx context.Context, client *awsssm.Client, cfgs []config.Config) ([]params.Pair, error) {
	perService := make([][]params.Parameter, 0, len(cfgs))
	for _, cfg := range cfgs {
		ps, err := client.GetParameters(ctx, cfg)
		if err != nil {
			return nil, err
		}
		perService = append(perService, ps)
	}
	return params.Merge(perService), nil
}

// newClient builds the SSM client shared by all subcommands.
func newClient(ctx context.Context) (*awsssm.Client, error) {
	return awsssm.New(ctx)
}
