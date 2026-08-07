package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/api"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/auth"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/config"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/keychain"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/version"
)

type agentAPI interface {
	RequestAgentOTP(ctx context.Context, request api.AgentOTPRequest, idempotencyKey string) (api.OTPRequestAccepted, error)
	VerifyAgentOTP(ctx context.Context, request api.AgentOTPVerificationRequest, idempotencyKey string) (api.AgentSessionResponse, error)
	CurrentUser(ctx context.Context, accessToken string) (api.CurrentAccount, error)
	ListRecorderCards(ctx context.Context, accessToken string) (api.AgentRecorderCardPage, error)
	GetQuotaSummary(ctx context.Context, accessToken string) (api.AgentQuotaSummary, error)
	GetModelUsageSummary(ctx context.Context, accessToken string) (api.AgentModelUsageSummary, error)
	ListMemories(ctx context.Context, accessToken string, params api.ListMemoriesParams) (api.AgentMemoryPage, error)
	GetMemory(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentMemory, error)
	Logout(ctx context.Context, accessToken string) error
}

type apiFactory func(config.Config) (agentAPI, error)

type runtimeState struct {
	Config config.Config
	Auth   *auth.Manager
	API    agentAPI
}

func loadRuntime(state *rootState) (runtimeState, error) {
	cfg, err := config.Load(state.viper, config.LoadOptions{
		GOOS:    state.targetOS,
		PathEnv: state.pathEnv,
	})
	if err != nil {
		return runtimeState{}, err
	}

	store := state.credentialStore
	if store == nil {
		store = defaultCredentialStore(cfg)
	}

	factory := state.apiFactory
	if factory == nil {
		factory = defaultAPIFactory
	}
	agentClient, err := factory(cfg)
	if err != nil {
		return runtimeState{}, err
	}

	return runtimeState{
		Config: cfg,
		Auth:   auth.NewManager(store, cfg.Profile),
		API:    agentClient,
	}, nil
}

func defaultCredentialStore(cfg config.Config) keychain.Store {
	if cfg.Auth.InsecureFileKeychain {
		return keychain.NewFileStore(filepath.Join(cfg.Paths.ConfigDir, "credentials.dev.json"))
	}
	return keychain.NewNativeStore()
}

func defaultAPIFactory(cfg config.Config) (agentAPI, error) {
	if strings.TrimSpace(cfg.Server.BaseURL) == "" {
		return nil, nil
	}
	client, err := api.New(api.Options{
		BaseURL:   cfg.Server.BaseURL,
		UserAgent: fmt.Sprintf("patchnote-agent/%s", version.Version),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
