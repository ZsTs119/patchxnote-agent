package keychain

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu         sync.Mutex
	credential map[string]Credential
	secrets    map[string]map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		credential: make(map[string]Credential),
		secrets:    make(map[string]map[string]string),
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

func (s *MemoryStore) GetSecret(ctx context.Context, profile string, name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	values, ok := s.secrets[profile]
	if !ok {
		return "", ErrNotFound
	}
	value, ok := values[name]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *MemoryStore) PutSecret(ctx context.Context, profile string, name string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.secrets[profile] == nil {
		s.secrets[profile] = make(map[string]string)
	}
	s.secrets[profile][name] = value
	return nil
}

func (s *MemoryStore) DeleteSecret(ctx context.Context, profile string, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.secrets[profile] != nil {
		delete(s.secrets[profile], name)
		if len(s.secrets[profile]) == 0 {
			delete(s.secrets, profile)
		}
	}
	return nil
}

func cloneCredential(credential Credential) Credential {
	cloned := credential
	cloned.Scopes = append([]string(nil), credential.Scopes...)
	return cloned
}
