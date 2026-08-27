package oauthflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
)

const (
	defaultRefreshWindow = 2 * time.Minute
	defaultLockWait      = 10 * time.Second
	defaultLockStale     = 2 * time.Minute
)

type TokenClient interface {
	RefreshOAuthToken(ctx context.Context, request api.OAuthTokenRequest) (api.OAuthTokenResponse, error)
}

type RefreshProvider struct {
	Store         *Store
	Client        TokenClient
	ServerBaseURL string
	ClientID      string
	RefreshWindow time.Duration
	LockPath      string
	Now           func() time.Time
	Sleep         func(context.Context, time.Duration) error
}

func (p *RefreshProvider) Credential(ctx context.Context) (Credential, bool, error) {
	credential, ok, err := p.loadMatching(ctx)
	if err != nil || !ok {
		return credential, ok, err
	}
	if !p.shouldRefresh(credential) {
		return credential, true, nil
	}
	return p.RefreshNow(ctx)
}

func (p *RefreshProvider) AccessToken(ctx context.Context) (string, bool, error) {
	credential, ok, err := p.Credential(ctx)
	if err != nil || !ok {
		return "", ok, err
	}
	return credential.AccessToken, true, nil
}

func (p *RefreshProvider) RefreshNow(ctx context.Context) (Credential, bool, error) {
	if p.Store == nil || p.Client == nil {
		return Credential{}, false, nil
	}
	release, err := p.acquireLock(ctx)
	if err != nil {
		return Credential{}, false, err
	}
	defer release()

	credential, ok, err := p.loadMatching(ctx)
	if err != nil || !ok {
		return credential, ok, err
	}
	now := p.now()
	if !credential.RefreshValid(now) {
		return Credential{}, false, nil
	}
	response, err := p.Client.RefreshOAuthToken(ctx, api.OAuthTokenRequest{
		ClientID:     p.clientID(),
		RefreshToken: credential.RefreshToken,
	})
	if err != nil {
		var apiErr *api.Error
		if errors.As(err, &apiErr) && apiErr.StatusCode == 401 {
			_ = p.Store.Delete(ctx)
		}
		return Credential{}, false, err
	}
	updated, err := CredentialFromTokenResponse(response, p.ServerBaseURL, p.clientID(), now)
	if err != nil {
		return Credential{}, false, err
	}
	if err := p.Store.Save(ctx, updated); err != nil {
		return Credential{}, false, err
	}
	return updated, true, nil
}

func (p *RefreshProvider) loadMatching(ctx context.Context) (Credential, bool, error) {
	if p.Store == nil {
		return Credential{}, false, nil
	}
	credential, ok, err := p.Store.Load(ctx)
	if err != nil || !ok {
		return credential, ok, err
	}
	if !credential.Matches(p.ServerBaseURL, p.clientID()) {
		return Credential{}, false, nil
	}
	return credential, true, nil
}

func (p *RefreshProvider) shouldRefresh(credential Credential) bool {
	if credential.AccessToken == "" {
		return credential.RefreshToken != ""
	}
	if credential.Metadata.AccessTokenExpiresAt.IsZero() {
		return false
	}
	return !p.now().Add(p.refreshWindow()).Before(credential.Metadata.AccessTokenExpiresAt)
}

func (p *RefreshProvider) acquireLock(ctx context.Context) (func(), error) {
	if p.LockPath == "" {
		return func() {}, nil
	}
	deadline := p.now().Add(defaultLockWait)
	for {
		if err := os.MkdirAll(filepath.Dir(p.LockPath), 0o700); err != nil {
			return nil, fmt.Errorf("create mcp oauth refresh lock directory: %w", err)
		}
		file, err := os.OpenFile(p.LockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
			return func() {
				_ = file.Close()
				_ = os.Remove(p.LockPath)
			}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("create mcp oauth refresh lock: %w", err)
		}
		if stale, statErr := p.lockIsStale(); statErr == nil && stale {
			_ = os.Remove(p.LockPath)
			continue
		}
		if !p.now().Before(deadline) {
			return nil, fmt.Errorf("timed out waiting for mcp oauth credential refresh lock")
		}
		if err := p.sleep(ctx, 100*time.Millisecond); err != nil {
			return nil, err
		}
	}
}

func (p *RefreshProvider) lockIsStale() (bool, error) {
	info, err := os.Stat(p.LockPath)
	if err != nil {
		return false, err
	}
	return p.now().Sub(info.ModTime()) > defaultLockStale, nil
}

func (p *RefreshProvider) refreshWindow() time.Duration {
	if p.RefreshWindow <= 0 {
		return defaultRefreshWindow
	}
	return p.RefreshWindow
}

func (p *RefreshProvider) clientID() string {
	if p.ClientID == "" {
		return DefaultClientID
	}
	return p.ClientID
}

func (p *RefreshProvider) now() time.Time {
	if p.Now != nil {
		return p.Now().UTC()
	}
	return time.Now().UTC()
}

func (p *RefreshProvider) sleep(ctx context.Context, duration time.Duration) error {
	if p.Sleep != nil {
		return p.Sleep(ctx, duration)
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
