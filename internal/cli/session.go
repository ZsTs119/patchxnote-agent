package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

const (
	defaultRefreshWindow = 2 * time.Minute
	defaultLockWait      = 10 * time.Second
	defaultLockStale     = 2 * time.Minute
)

type sessionCredentialProvider struct {
	auth          *auth.Manager
	api           agentAPI
	lockPath      string
	refreshWindow time.Duration
	lockWait      time.Duration
	lockStale     time.Duration
	mu            sync.Mutex
	now           func() time.Time
	sleep         func(context.Context, time.Duration) error
}

func newSessionCredentialProvider(manager *auth.Manager, agentClient agentAPI, cfg config.Config) *sessionCredentialProvider {
	return &sessionCredentialProvider{
		auth:          manager,
		api:           agentClient,
		lockPath:      filepath.Join(cfg.Paths.ConfigDir, "agent-refresh.lock"),
		refreshWindow: defaultRefreshWindow,
		lockWait:      defaultLockWait,
		lockStale:     defaultLockStale,
		now:           time.Now,
		sleep:         sleepContext,
	}
}

func (p *sessionCredentialProvider) Credential(ctx context.Context) (keychain.Credential, bool, error) {
	credential, ok, err := p.auth.Credential(ctx)
	if err != nil || !ok {
		return credential, ok, err
	}
	if !p.shouldRefresh(credential) {
		return credential, true, nil
	}
	if credential.RefreshToken == "" || p.api == nil {
		return keychain.Credential{}, false, nil
	}

	return p.refresh(ctx, false)
}

func (p *sessionCredentialProvider) RefreshNow(ctx context.Context) (keychain.Credential, bool, error) {
	return p.refresh(ctx, true)
}

func (p *sessionCredentialProvider) refresh(ctx context.Context, force bool) (keychain.Credential, bool, error) {
	if p.api == nil {
		return keychain.Credential{}, false, nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	release, err := p.acquireLock(ctx)
	if err != nil {
		return keychain.Credential{}, false, err
	}
	defer release()

	credential, ok, err := p.auth.Credential(ctx)
	if err != nil || !ok {
		return credential, ok, err
	}
	if !force && !p.shouldRefresh(credential) && credential.AccessToken != "" {
		return credential, true, nil
	}
	if credential.RefreshToken == "" {
		return keychain.Credential{}, false, nil
	}

	idempotencyKey, err := newOpaqueID("idem")
	if err != nil {
		return keychain.Credential{}, false, err
	}
	session, err := p.api.RefreshAgentSession(ctx, api.AgentRefreshRequest{RefreshToken: credential.RefreshToken}, idempotencyKey)
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == 401 {
			_ = p.auth.Logout(ctx)
		}
		return keychain.Credential{}, false, err
	}
	if session.AccessToken == "" || session.RefreshToken == "" {
		return keychain.Credential{}, false, fmt.Errorf("agent refresh response did not include rotated credentials")
	}

	now := p.now()
	updated := keychain.Credential{
		AccountID:             session.Account.ID,
		AccessToken:           session.AccessToken,
		RefreshToken:          session.RefreshToken,
		AccessTokenExpiresAt:  now.Add(time.Duration(session.AccessExpiresInSeconds) * time.Second),
		RefreshTokenExpiresAt: now.Add(time.Duration(session.RefreshExpiresInSeconds) * time.Second),
		Scopes:                append([]string(nil), session.Scopes...),
	}
	if err := p.auth.Save(ctx, updated); err != nil {
		return keychain.Credential{}, false, err
	}
	return updated, true, nil
}

func (p *sessionCredentialProvider) shouldRefresh(credential keychain.Credential) bool {
	if credential.AccessToken == "" {
		return credential.RefreshToken != ""
	}
	if credential.AccessTokenExpiresAt.IsZero() {
		return false
	}
	return !p.now().Add(p.refreshWindow).Before(credential.AccessTokenExpiresAt)
}

func (p *sessionCredentialProvider) acquireLock(ctx context.Context) (func(), error) {
	if p.lockPath == "" {
		return func() {}, nil
	}
	deadline := p.now().Add(p.lockWait)
	for {
		if err := os.MkdirAll(filepath.Dir(p.lockPath), 0o700); err != nil {
			return nil, fmt.Errorf("create refresh lock directory: %w", err)
		}
		file, err := os.OpenFile(p.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			return func() {
				_ = file.Close()
				_ = os.Remove(p.lockPath)
			}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("create refresh lock: %w", err)
		}
		if stale, statErr := p.lockIsStale(); statErr == nil && stale {
			_ = os.Remove(p.lockPath)
			continue
		}
		if !p.now().Before(deadline) {
			return nil, fmt.Errorf("timed out waiting for agent credential refresh lock")
		}
		if err := p.sleep(ctx, 100*time.Millisecond); err != nil {
			return nil, err
		}
	}
}

func (p *sessionCredentialProvider) lockIsStale() (bool, error) {
	info, err := os.Stat(p.lockPath)
	if err != nil {
		return false, err
	}
	return p.now().Sub(info.ModTime()) > p.lockStale, nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
