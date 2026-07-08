package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newFileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "file <path>",
		Short: "Writes configuration to a file as `export KEY=VALUE` lines.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			path := args[0]

			// The parent directory must already exist; we do not create it.
			if dir := filepath.Dir(path); dir != "" {
				if _, err := os.Stat(dir); err != nil {
					return fmt.Errorf("directory %q does not exist", dir)
				}
			}

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

			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer f.Close()

			w := bufio.NewWriter(f)
			for _, p := range pairs {
				if _, err := fmt.Fprintf(w, "export %s=%s\n", p.Key, p.Value); err != nil {
					return err
				}
			}
			return w.Flush()
		},
	}
}
