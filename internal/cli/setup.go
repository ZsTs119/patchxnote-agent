package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/oauthflow"
	clientsetup "github.com/ZsTs119/patchxnote-agent/internal/setup"

	"github.com/spf13/cobra"
)

type setupSessionClient interface {
	CreateAgentSetupSession(ctx context.Context, request api.AgentSetupSessionCreateRequest, idempotencyKey string) (api.AgentSetupSessionCreated, error)
	GetAgentSetupSession(ctx context.Context, sessionID string) (api.AgentSetupSessionStatus, error)
}

type setupCommandResult struct {
	AuthStatus        string                              `json:"auth_status"`
	AuthMethod        string                              `json:"auth_method,omitempty"`
	Profile           string                              `json:"profile"`
	RuntimeOS         string                              `json:"runtime_os"`
	Clients           []clientsetup.InstallResult         `json:"clients"`
	MCPConfig         clientsetup.MCPServerConfig         `json:"mcp_config"`
	MCPProtocolSmoke  *clientsetup.MCPProtocolSmokeResult `json:"mcp_protocol_smoke,omitempty"`
	SafeAccountID     string                              `json:"safe_account_id,omitempty"`
	SetupSessionState string                              `json:"setup_session_state,omitempty"`
}

func newSetupCommand(state *rootState) *cobra.Command {
	var clientID string
	var allLocalSupported bool
	var dryRun bool
	var noBrowser bool
	var printConfig bool
	var force bool
	var yes bool
	var skipMCPSmoke bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up PatchXNote MCP in a supported local client",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtimeState, err := loadRuntime(state)
			if err != nil {
				return err
			}
			targetOS := state.targetOS
			if targetOS == "" {
				targetOS = runtime.GOOS
			}

			clientIDs, err := selectedSetupClients(clientID, allLocalSupported)
			if err != nil {
				return err
			}
			pathEnv := state.pathEnv
			if pathEnv == (config.PathEnv{}) {
				pathEnv = config.DefaultPathEnv()
			}
			plans := make([]clientsetup.ClientPlan, 0, len(clientIDs))
			for _, id := range clientIDs {
				plan, err := clientsetup.BuildPlan(id, clientsetup.PlanOptions{
					TargetOS: targetOS,
					PathEnv:  pathEnv,
				})
				if err != nil {
					return err
				}
				plans = append(plans, plan)
			}
			if _, err := exec.LookPath("npx"); err != nil {
				for index := range plans {
					plans[index].Warnings = append(plans[index].Warnings, "npx is not on PATH; install Node.js/npm in this same runtime or use the native patchxnote binary in the MCP command.")
				}
			}

			result := setupCommandResult{
				AuthStatus: "not_checked",
				Profile:    runtimeState.Config.Profile,
				RuntimeOS:  targetOS,
				MCPConfig:  clientsetup.DefaultServerConfig(),
			}

			format := normalizedOutputFormat(state)
			if format != "" && format != "plain" && format != "json" {
				return unsupportedOutputFormatError(format)
			}

			if !dryRun {
				method, accountID, err := ensureAuthenticatedForSetup(cmd, state, runtimeState, plans, noBrowser)
				if err != nil {
					return err
				}
				if method == "remote_platform_pending" {
					result.AuthStatus = "not_required"
				} else {
					result.AuthStatus = "authenticated"
				}
				result.AuthMethod = method
				result.SafeAccountID = accountID
			}

			reader := bufio.NewReader(cmd.InOrStdin())
			for _, plan := range plans {
				if format != "json" {
					printSetupPlan(cmd, plan, dryRun, printConfig)
				}
				if !dryRun && !plan.ManualRequired && !yes {
					ok, err := confirmConfigWrite(cmd, reader, plan)
					if err != nil {
						return err
					}
					if !ok {
						result.Clients = append(result.Clients, clientsetup.InstallResult{
							Status:         "manual_required",
							ClientID:       plan.Client.ID,
							ConfigPath:     plan.ConfigPath,
							ManualRequired: true,
							ManualReason:   "User skipped automatic config write.",
							Warnings:       plan.Warnings,
						})
						continue
					}
				}
				applied, err := clientsetup.ApplyPlan(plan, clientsetup.InstallOptions{DryRun: dryRun, Force: force})
				if err != nil {
					return err
				}
				result.Clients = append(result.Clients, applied)
				if format != "json" {
					printSetupResult(cmd, plan, applied)
				}
			}

			if !dryRun && !skipMCPSmoke {
				smoke, err := runSetupMCPSmoke(cmd.Context())
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "MCP smoke skipped or failed: %v\n", err)
				} else {
					result.MCPProtocolSmoke = &smoke
					if format != "json" {
						fmt.Fprintf(cmd.OutOrStdout(), "MCP smoke: initialize ok, tools/list ok (%d tools)\n", smoke.ToolCount)
					}
				}
			}

			if format == "json" {
				return writeJSON(cmd.OutOrStdout(), result)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&clientID, "client", "", "Client id to set up")
	cmd.Flags().BoolVar(&allLocalSupported, "all-local-supported", false, "Set up all locally auto-writable P0 clients sequentially")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show planned setup actions without modifying client config")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Do not open a browser; print login guidance instead")
	cmd.Flags().BoolVar(&printConfig, "print-config", false, "Print the client MCP config snippet")
	cmd.Flags().BoolVar(&force, "force", false, "Replace only an existing patchxnote MCP entry when it differs")
	cmd.Flags().BoolVar(&yes, "yes", false, "Apply config changes without interactive confirmation")
	cmd.Flags().BoolVar(&skipMCPSmoke, "skip-mcp-smoke", false, "Skip direct MCP protocol smoke after setup")
	return cmd
}

func selectedSetupClients(clientID string, allLocalSupported bool) ([]string, error) {
	clientID = strings.TrimSpace(clientID)
	if allLocalSupported {
		if clientID != "" {
			return nil, fmt.Errorf("--client cannot be combined with --all-local-supported")
		}
		return clientsetup.LocalSupportedClientIDs(), nil
	}
	if clientID == "" {
		return nil, fmt.Errorf("--client is required unless --all-local-supported is set")
	}
	return []string{clientID}, nil
}

func ensureAuthenticatedForSetup(cmd *cobra.Command, state *rootState, runtimeState runtimeState, plans []clientsetup.ClientPlan, noBrowser bool) (string, string, error) {
	if !setupNeedsLocalMCPAuth(plans) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Hosted remote MCP setup is platform-side; local browser login was not started.")
		return "remote_platform_pending", "", nil
	}
	if runtimeState.API == nil {
		return "", "", fmt.Errorf("server base URL is required; set --server-base-url or PATCHXNOTE_SERVER_BASE_URL")
	}
	store := newMCPOAuthStore(runtimeState)
	provider := newMCPOAuthRefreshProvider(runtimeState, store, oauthflow.DefaultClientID)
	credential, ok, err := provider.Credential(cmd.Context())
	if err != nil {
		if errors.Is(err, oauthflow.ErrInvalidCredential) {
			return "", "", fmt.Errorf("stored MCP OAuth credential is invalid; run patchxnote mcp logout --local-only, then patchxnote mcp login")
		}
		return "", "", err
	}
	if ok && credential.RefreshValid(time.Now().UTC()) {
		return "mcp_oauth", "", nil
	}
	result, err := runMCPLoginFlow(cmd, state, runtimeState, mcpLoginOptions{
		NoBrowser:       noBrowser,
		Force:           false,
		SkipSmoke:       true,
		CallbackTimeout: defaultMCPLoginTimeout,
		ClientID:        oauthflow.DefaultClientID,
	})
	if err != nil {
		return "", "", err
	}
	if !result.LoggedIn {
		return "", "", fmt.Errorf("mcp oauth login did not complete")
	}
	return "mcp_oauth", "", nil
}

func setupNeedsLocalMCPAuth(plans []clientsetup.ClientPlan) bool {
	for _, plan := range plans {
		if plan.Client.PrimaryStrategy != "remote-url" {
			return true
		}
	}
	return false
}

// Historical setup-session path kept only for compatibility tests and quick rollback context.
func tryBrowserSetupSession(cmd *cobra.Command, runtimeState runtimeState, plan clientsetup.ClientPlan) (string, string, bool, error) {
	client, ok := runtimeState.API.(setupSessionClient)
	if !ok {
		return "", "", false, nil
	}
	idempotencyKey, err := newOpaqueID("idem")
	if err != nil {
		return "", "", false, err
	}
	created, err := client.CreateAgentSetupSession(cmd.Context(), api.AgentSetupSessionCreateRequest{
		ClientID:   plan.Client.ID,
		ClientName: plan.Client.Name,
		Profile:    runtimeState.Config.Profile,
		Scopes:     []string{"agent:account.read", "agent:memories.read"},
	}, idempotencyKey)
	if err != nil {
		return "", "", false, err
	}
	loginURL := created.VerificationURIFull
	if loginURL == "" {
		loginURL = created.VerificationURI
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Open PatchXNote login: %s\n", loginURL)
	if created.UserCode != "" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Setup code: %s\n", created.UserCode)
	}
	if err := openBrowserBestEffort(loginURL); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Could not open browser automatically: %v\n", err)
	}
	session, err := pollSetupSession(cmd.Context(), client, created)
	if err != nil {
		return "", "", false, err
	}
	if session == nil {
		return "", "", false, nil
	}
	if err := saveAgentSession(cmd.Context(), runtimeState, *session); err != nil {
		return "", "", false, err
	}
	return "browser_setup_session", session.Account.ID, true, nil
}

func pollSetupSession(ctx context.Context, client setupSessionClient, created api.AgentSetupSessionCreated) (*api.AgentSessionResponse, error) {
	timeout := time.Duration(created.ExpiresInSeconds) * time.Second
	if timeout <= 0 || timeout > 10*time.Minute {
		timeout = 10 * time.Minute
	}
	interval := time.Duration(created.PollIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if interval > 5*time.Second {
		interval = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		status, err := client.GetAgentSetupSession(ctx, created.SessionID)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(strings.TrimSpace(status.Status)) {
		case "approved", "completed", "succeeded":
			if status.Session == nil {
				return nil, fmt.Errorf("setup session approved without credentials")
			}
			return status.Session, nil
		case "denied", "rejected", "expired":
			return nil, fmt.Errorf("setup session %s", status.Status)
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("setup session timed out")
		case <-timer.C:
		}
	}
}

func runSetupOTPLogin(cmd *cobra.Command, runtimeState runtimeState) (string, error) {
	reader := bufio.NewReader(cmd.InOrStdin())
	phone, err := readPromptLine(cmd.ErrOrStderr(), reader, "Phone: ")
	if err != nil {
		return "", err
	}
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "", fmt.Errorf("phone is required")
	}
	clientInstance, err := newOpaqueID("agent_cli")
	if err != nil {
		return "", err
	}
	requestIDKey, err := newOpaqueID("idem")
	if err != nil {
		return "", err
	}
	accepted, err := runtimeState.API.RequestAgentOTP(cmd.Context(), api.AgentOTPRequest{
		Phone:          phone,
		ClientInstance: clientInstance,
	}, requestIDKey)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Verification code sent. Retry after %d seconds if needed.\n", accepted.CooldownSeconds)

	code, err := readSecretPromptLine(cmd, reader, "Verification code: ")
	if err != nil {
		return "", err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("verification code is required")
	}
	verifyIDKey, err := newOpaqueID("idem")
	if err != nil {
		return "", err
	}
	session, err := runtimeState.API.VerifyAgentOTP(cmd.Context(), api.AgentOTPVerificationRequest{
		RequestID:      accepted.RequestID,
		Code:           code,
		ClientInstance: clientInstance,
	}, verifyIDKey)
	if err != nil {
		return "", err
	}
	if err := saveAgentSession(cmd.Context(), runtimeState, session); err != nil {
		return "", err
	}
	return session.Account.ID, nil
}

func saveAgentSession(ctx context.Context, runtimeState runtimeState, session api.AgentSessionResponse) error {
	credential := keychain.Credential{
		AccountID:             session.Account.ID,
		AccessToken:           session.AccessToken,
		RefreshToken:          session.RefreshToken,
		AccessTokenExpiresAt:  time.Now().Add(time.Duration(session.AccessExpiresInSeconds) * time.Second),
		RefreshTokenExpiresAt: time.Now().Add(time.Duration(session.RefreshExpiresInSeconds) * time.Second),
		Scopes:                append([]string(nil), session.Scopes...),
	}
	return runtimeState.Auth.Save(ctx, credential)
}

func isUnsupportedSetupSessionError(err error) bool {
	var apiErr *api.Error
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusNotFound || apiErr.Code == "route_not_found" || apiErr.Code == "not_found"
}

func openBrowserBestEffort(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("login URL is empty")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	case "darwin":
		cmd = exec.Command("open", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}

func confirmConfigWrite(cmd *cobra.Command, reader *bufio.Reader, plan clientsetup.ClientPlan) (bool, error) {
	fmt.Fprintf(cmd.ErrOrStderr(), "Modify %s for %s? [y/N] ", plan.ConfigPath, plan.Client.Name)
	answer, err := reader.ReadString('\n')
	if err != nil && answer == "" {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func printSetupPlan(cmd *cobra.Command, plan clientsetup.ClientPlan, dryRun bool, printConfig bool) {
	if dryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "Dry run: %s (%s)\n", plan.Client.Name, plan.Client.ID)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "Setup: %s (%s)\n", plan.Client.Name, plan.Client.ID)
	}
	if plan.ConfigPath != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n", plan.ConfigPath)
	}
	if plan.CodexCommand != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Codex command: %s\n", plan.CodexCommand)
	}
	if plan.ClaudeCommand != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Claude command: %s\n", plan.ClaudeCommand)
	}
	if plan.Deeplink != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Install link: %s\n", plan.Deeplink)
	}
	if printConfig || plan.ManualRequired {
		fmt.Fprintln(cmd.OutOrStdout(), "MCP config:")
		fmt.Fprintln(cmd.OutOrStdout(), plan.ConfigSnippet)
	}
	for _, warning := range plan.Warnings {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %s\n", warning)
	}
	if plan.ManualRequired {
		fmt.Fprintf(cmd.OutOrStdout(), "Manual setup required: %s\n", plan.ManualReason)
		fmt.Fprintf(cmd.OutOrStdout(), "After adding the server, ask the client to list PatchXNote MCP tools.\n")
	}
}

func printSetupResult(cmd *cobra.Command, plan clientsetup.ClientPlan, result clientsetup.InstallResult) {
	switch result.Status {
	case "dry_run":
		fmt.Fprintf(cmd.OutOrStdout(), "No files changed for %s.\n", plan.Client.ID)
	case "installed":
		if result.Changed {
			fmt.Fprintf(cmd.OutOrStdout(), "Installed PatchXNote MCP config for %s.\n", plan.Client.ID)
			if result.BackupPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Backup: %s\n", result.BackupPath)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "PatchXNote MCP config already exists for %s.\n", plan.Client.ID)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Next: %s\n", plan.Client.RequiresRestart)
	case "manual_required":
		fmt.Fprintf(cmd.OutOrStdout(), "Manual path kept for %s.\n", plan.Client.ID)
	}
}

func runSetupMCPSmoke(ctx context.Context) (clientsetup.MCPProtocolSmokeResult, error) {
	executable, err := os.Executable()
	if err != nil {
		return clientsetup.MCPProtocolSmokeResult{}, err
	}
	return clientsetup.SmokeMCPCommand(ctx, executable, []string{"mcp", "serve"}, 10*time.Second)
}
