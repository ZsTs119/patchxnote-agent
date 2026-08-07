package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogoutCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove local PatchXNote Agent credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}

			credential, ok, err := runtime.Credentials.Credential(cmd.Context())
			if err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning: server logout failed; local credentials will still be removed")
				ok = false
			}
			if ok && runtime.API != nil && credential.AccessToken != "" {
				if err := runtime.API.Logout(cmd.Context(), credential.AccessToken); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "warning: server logout failed; local credentials will still be removed")
				}
			}

			if err := runtime.Auth.Logout(cmd.Context()); err != nil {
				return err
			}

			result := struct {
				LoggedOut bool   `json:"logged_out"`
				Profile   string `json:"profile"`
			}{
				LoggedOut: true,
				Profile:   runtime.Config.Profile,
			}

			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "logged out\nprofile %s\n", result.Profile)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
}
