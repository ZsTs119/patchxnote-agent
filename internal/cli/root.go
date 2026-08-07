package cli

import (
	"context"
	"io"
	"os"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Deps struct {
	TargetOS        string
	PathEnv         config.PathEnv
	CredentialStore keychain.Store
	APIFactory      apiFactory
	Stdin           io.Reader
}

type rootState struct {
	viper           *viper.Viper
	targetOS        string
	pathEnv         config.PathEnv
	credentialStore keychain.Store
	apiFactory      apiFactory
}

func Execute() error {
	return ExecuteContext(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
}

func ExecuteContext(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	cmd := NewRootCommand()
	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd.Execute()
}

func NewRootCommand() *cobra.Command {
	return NewRootCommandWithDeps(Deps{})
}

func NewRootCommandWithDeps(deps Deps) *cobra.Command {
	state := &rootState{
		viper:           config.NewViper(),
		targetOS:        deps.TargetOS,
		pathEnv:         deps.PathEnv,
		credentialStore: deps.CredentialStore,
		apiFactory:      deps.APIFactory,
	}

	cmd := &cobra.Command{
		Use:               "patchxnote",
		Short:             "PatchXNote Agent CLI and local MCP bridge",
		SilenceUsage:      true,
		SilenceErrors:     true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	flags := cmd.PersistentFlags()
	flags.String("config", "", "Path to a non-secret config file")
	flags.String("profile", "default", "Config profile name")
	flags.StringP("output", "o", "plain", "Output format: plain or json")
	flags.String("server-base-url", "", "PatchXNote API base URL; defaults to the PatchXNote test API")

	mustBind(state.viper, "config", flags.Lookup("config"))
	mustBind(state.viper, "profile", flags.Lookup("profile"))
	mustBind(state.viper, "output", flags.Lookup("output"))
	mustBind(state.viper, "server.base_url", flags.Lookup("server-base-url"))

	cmd.AddCommand(
		newAuthCommand(state),
		newLoginCommand(state),
		newLogoutCommand(state),
		newMCPCommand(state),
		newVersionCommand(state),
	)
	if deps.Stdin != nil {
		cmd.SetIn(deps.Stdin)
	}

	return cmd
}

func mustBind(v *viper.Viper, key string, flag *pflag.Flag) {
	if err := v.BindPFlag(key, flag); err != nil {
		panic(err)
	}
}
