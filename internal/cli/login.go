package cli

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newLoginCommand(state *rootState) *cobra.Command {
	var phone string
	var code string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to PatchXNote Agent with phone OTP",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			if runtime.API == nil {
				return fmt.Errorf("server base URL is required; set --server-base-url or PATCHXNOTE_SERVER_BASE_URL")
			}

			reader := bufio.NewReader(cmd.InOrStdin())
			if strings.TrimSpace(phone) == "" {
				phone, err = readPromptLine(cmd.ErrOrStderr(), reader, "Phone: ")
				if err != nil {
					return err
				}
			}
			phone = strings.TrimSpace(phone)
			if phone == "" {
				return fmt.Errorf("phone is required")
			}

			clientInstance, err := newOpaqueID("agent_cli")
			if err != nil {
				return err
			}
			requestIDKey, err := newOpaqueID("idem")
			if err != nil {
				return err
			}
			accepted, err := runtime.API.RequestAgentOTP(cmd.Context(), api.AgentOTPRequest{
				Phone:          phone,
				ClientInstance: clientInstance,
			}, requestIDKey)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Verification code sent. Retry after %d seconds if needed.\n", accepted.CooldownSeconds)

			if strings.TrimSpace(code) == "" {
				code, err = readSecretPromptLine(cmd, reader, "Verification code: ")
				if err != nil {
					return err
				}
			}
			code = strings.TrimSpace(code)
			if code == "" {
				return fmt.Errorf("verification code is required")
			}

			verifyIDKey, err := newOpaqueID("idem")
			if err != nil {
				return err
			}
			session, err := runtime.API.VerifyAgentOTP(cmd.Context(), api.AgentOTPVerificationRequest{
				RequestID:      accepted.RequestID,
				Code:           code,
				ClientInstance: clientInstance,
			}, verifyIDKey)
			if err != nil {
				return err
			}

			credential := keychain.Credential{
				AccountID:            session.Account.ID,
				AccessToken:          session.AccessToken,
				RefreshToken:         session.RefreshToken,
				AccessTokenExpiresAt: time.Now().Add(time.Duration(session.AccessExpiresInSeconds) * time.Second),
				Scopes:               session.Scopes,
			}
			if err := runtime.Auth.Save(cmd.Context(), credential); err != nil {
				return err
			}

			result := struct {
				LoggedIn bool               `json:"logged_in"`
				Profile  string             `json:"profile"`
				Account  api.CurrentAccount `json:"account"`
				Scopes   []string           `json:"scopes"`
			}{
				LoggedIn: true,
				Profile:  runtime.Config.Profile,
				Account:  session.Account,
				Scopes:   session.Scopes,
			}

			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "login succeeded\nprofile %s\naccount %s\n", result.Profile, result.Account.ID)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}

	cmd.Flags().StringVar(&phone, "phone", "", "Phone number for OTP login")
	cmd.Flags().StringVar(&code, "code", "", "Six digit OTP code")
	return cmd
}

func readPromptLine(stderr io.Writer, reader *bufio.Reader, prompt string) (string, error) {
	fmt.Fprint(stderr, prompt)
	line, err := reader.ReadString('\n')
	if err != nil && !(err == io.EOF && line != "") {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func readSecretPromptLine(cmd *cobra.Command, reader *bufio.Reader, prompt string) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), prompt)
	if file, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		value, err := term.ReadPassword(int(file.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(value)), nil
	}
	line, err := reader.ReadString('\n')
	if err != nil && !(err == io.EOF && line != "") {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func newOpaqueID(prefix string) (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate opaque id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(bytes[:]), nil
}
