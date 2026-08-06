package cli

import (
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/cache"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/mcp"

	"github.com/spf13/cobra"
)

func newMCPCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run or configure the PatchNote MCP server",
	}
	cmd.AddCommand(newMCPServeCommand(state))
	return cmd
}

func newMCPServeCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Serve PatchNote MCP over stdio",
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
			})
			return server.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}
