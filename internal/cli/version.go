package cli

import (
	"fmt"

	"github.com/ZsTs119/patchxnote-agent/internal/version"

	"github.com/spf13/cobra"
)

func newVersionCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print PatchXNote Agent version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Current()

			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err := fmt.Fprintf(cmd.OutOrStdout(),
					"patchxnote %s\ncommit %s\ndate %s\ngo %s\nplatform %s/%s\n",
					info.Version,
					info.Commit,
					info.Date,
					info.GoVersion,
					info.OS,
					info.Arch,
				)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), info)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
}
