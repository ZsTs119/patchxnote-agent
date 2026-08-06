package cli

import (
	"fmt"

	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/version"

	"github.com/spf13/cobra"
)

func newVersionCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print PatchNote Agent version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Current()

			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err := fmt.Fprintf(cmd.OutOrStdout(),
					"patchnote %s\ncommit %s\ndate %s\ngo %s\nplatform %s/%s\n",
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
