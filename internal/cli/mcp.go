package cli

import (
	"fmt"

	"github.com/ZsTs119/patchxnote-agent/internal/cache"
	"github.com/ZsTs119/patchxnote-agent/internal/mcp"
	"github.com/ZsTs119/patchxnote-agent/internal/remotemcp"
	"github.com/ZsTs119/patchxnote-agent/internal/version"

	"github.com/spf13/cobra"
)

func newMCPCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Run or configure the PatchXNote MCP server",
	}
	cmd.AddCommand(
		newMCPConfigCommand(state),
		newMCPLoginCommand(state),
		newMCPStatusCommand(state),
		newMCPLogoutCommand(state),
		newMCPServeCommand(state),
	)
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
			store := newMCPOAuthStore(runtime)
			clientID := normalizedMCPClientID("")
			useRemote, err := shouldUseRemoteMCP(cmd.Context(), runtime, store, clientID)
			if err != nil {
				return err
			}
			if useRemote {
				client, err := remotemcp.New(remotemcp.Options{
					ServerBaseURL: runtime.Config.Server.BaseURL,
					UserAgent:     fmt.Sprintf("patchxnote-agent/%s", version.Version),
				})
				if err != nil {
					return err
				}
				provider := newMCPOAuthRefreshProvider(runtime, store, clientID)
				proxy, err := remotemcp.NewProxy(remotemcp.ProxyOptions{
					Client:        client,
					TokenProvider: mcpOAuthRemoteTokenProvider{provider: provider},
				})
				if err != nil {
					return err
				}
				return proxy.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
			}
			server := mcp.NewServer(mcp.Options{
				Authenticator: runtime.Auth,
				Credentials:   runtime.Credentials,
				API:           runtime.API,
				MemoryCache:   cache.NewMemoryIndex(),
				Config:        runtime.Config,
				Secrets:       runtime.Secrets,
				Version:       version.Version,
			})
			return server.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}
