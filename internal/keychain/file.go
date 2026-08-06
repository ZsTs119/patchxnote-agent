package keychain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStore struct {
	mu   sync.Mutex
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Get(ctx context.Context, profile string) (Credential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	credentials, err := s.read()
	if err != nil {
		return Credential{}, err
	}
	credential, ok := credentials[profile]
	if !ok {
		return Credential{}, ErrNotFound
	}
	return cloneCredential(credential), nil
}

func (s *FileStore) Put(ctx context.Context, profile string, credential Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	credentials, err := s.read()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if credentials == nil {
		credentials = make(map[string]Credential)
	}
	credentials[profile] = cloneCredential(credential)
	return s.write(credentials)
}

func (s *FileStore) Delete(ctx context.Context, profile string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	credentials, err := s.read()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	delete(credentials, profile)
	return s.write(credentials)
}

func (s *FileStore) read() (map[string]Credential, error) {
	body, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read credential store: %w", err)
	}
	if len(body) == 0 {
		return map[string]Credential{}, nil
	}
	var credentials map[string]Credential
	if err := json.Unmarshal(body, &credentials); err != nil {
		return nil, fmt.Errorf("decode credential store: %w", err)
	}
	return credentials, nil
}

func (s *FileStore) write(credentials map[string]Credential) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create credential store dir: %w", err)
	}
	body, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credential store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return fmt.Errorf("write credential store: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace credential store: %w", err)
	}
	return nil
}
