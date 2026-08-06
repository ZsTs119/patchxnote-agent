package keychain

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu         sync.Mutex
	credential map[string]Credential
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		credential: make(map[string]Credential),
	}
}

func (s *MemoryStore) Get(ctx context.Context, profile string) (Credential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	credential, ok := s.credential[profile]
	if !ok {
		return Credential{}, ErrNotFound
	}
	return cloneCredential(credential), nil
}

func (s *MemoryStore) Put(ctx context.Context, profile string, credential Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.credential[profile] = cloneCredential(credential)
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, profile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.credential, profile)
	return nil
}

func cloneCredential(credential Credential) Credential {
	cloned := credential
	cloned.Scopes = append([]string(nil), credential.Scopes...)
	return cloned
}
