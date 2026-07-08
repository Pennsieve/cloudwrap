package cmd

import (
	"sort"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/pennsieve/cloudwrap/internal/params"
)

func newDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Prints configuration keys and metadata (no values).",
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

			// Union metadata across services, first occurrence of a key wins.
			seen := make(map[string]struct{})
			var rows []params.Metadata
			for _, cfg := range cfgs {
				md, err := client.DescribeParameters(ctx, cfg)
				if err != nil {
					return err
				}
				for _, m := range md {
					k := params.ShortName(m.Name)
					if _, dup := seen[k]; dup {
						continue
					}
					seen[k] = struct{}{}
					rows = append(rows, m)
				}
			}

			sort.Slice(rows, func(i, j int) bool {
				return params.ShortName(rows[i].Name) < params.ShortName(rows[j].Name)
			})

			table := tablewriter.NewWriter(cmd.OutOrStdout())
			table.SetHeader([]string{"KEY", "VERSION", "LAST_MODIFIED_USER", "LAST_MODIFIED_DATE"})
			for _, m := range rows {
				date := ""
				if !m.LastModifiedDate.IsZero() {
					date = m.LastModifiedDate.UTC().Format("2006-01-02 15:04:05")
				}
				table.Append([]string{
					params.ShortName(m.Name),
					strconv.FormatInt(m.Version, 10),
					m.LastModifiedUser,
					date,
				})
			}
			table.Render()
			return nil
		},
	}
}
