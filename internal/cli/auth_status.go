package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Inspect PatchNote Agent authentication",
	}

	cmd.AddCommand(newAuthStatusCommand(state))
	return cmd
}

func newAuthStatusCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print local authentication status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}

			status, err := runtime.Auth.Status(cmd.Context())
			if err != nil {
				return err
			}
			if status.Authenticated && runtime.API != nil {
				credential, ok, err := runtime.Auth.Credential(cmd.Context())
				if err != nil {
					return err
				}
				if ok && credential.AccessToken != "" {
					account, err := runtime.API.CurrentUser(cmd.Context(), credential.AccessToken)
					if err != nil {
						return err
					}
					status.AccountID = account.ID
					status.AccountStatus = account.Status
					status.RegistrationPlatform = account.RegistrationPlatform
					status.PhoneMasked = account.PhoneMasked
					status.StateVersion = account.StateVersion
				}
			}

			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				if status.Authenticated {
					_, err = fmt.Fprintf(cmd.OutOrStdout(), "authenticated\nprofile %s\naccount %s\n", status.Profile, status.AccountID)
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "unauthenticated\nprofile %s\n", status.Profile)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), status)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
}
