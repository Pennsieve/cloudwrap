package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStdoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stdout",
		Short: "Prints configuration to stdout as KEY=VALUE lines.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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

			for _, p := range pairs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", p.Key, p.Value)
			}
			return nil
		},
	}
}
