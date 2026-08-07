package cli

import (
	"github.com/ZsTs119/patchxnote-agent/internal/cache"
	"github.com/ZsTs119/patchxnote-agent/internal/mcp"
	"github.com/ZsTs119/patchxnote-agent/internal/version"

	"github.com/spf13/cobra"
)

func newMCPCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run or configure the PatchXNote MCP server",
	}
	cmd.AddCommand(newMCPServeCommand(state))
	return cmd
}

func newMCPServeCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve PatchXNote MCP over stdio",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			server := mcp.NewServer(mcp.Options{
				Authenticator: runtime.Auth,
				Credentials:   runtime.Auth,
				API:           runtime.API,
				MemoryCache:   cache.NewMemoryIndex(),
				Version:       version.Version,
			})
			return server.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}
