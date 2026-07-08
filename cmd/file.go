package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pennsieve/cloudwrap/internal/params"
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

			return writeExports(path, pairs)
		},
	}
}

// writeExports writes the pairs to path as `export KEY=VALUE` lines. If any
// write fails, the incomplete file is closed and removed so no partial output
// is left on disk.
func writeExports(path string, pairs []params.Pair) error {
	return writeExportsWith(path, pairs, writeExportLines)
}

// writeExportsWith owns the file lifecycle (create, and remove-on-failure) and
// delegates the actual byte writing to write. It exists as a seam so tests can
// inject a failing writer to exercise the cleanup path.
func writeExportsWith(path string, pairs []params.Pair, write func(io.Writer, []params.Pair) error) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	// On any failure, discard the partially written file.
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(path)
		}
	}()

	if err = write(f, pairs); err != nil {
		return err
	}
	return f.Close()
}

// writeExportLines writes each pair to w as a single `export KEY=VALUE` line.
func writeExportLines(w io.Writer, pairs []params.Pair) error {
	bw := bufio.NewWriter(w)
	for _, p := range pairs {
		if _, err := fmt.Fprintf(bw, "export %s=%s\n", p.Key, p.Value); err != nil {
			return err
		}
	}
	return bw.Flush()
}
