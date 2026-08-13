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

func (s *FileStore) GetSecret(ctx context.Context, profile string, name string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	secrets, err := s.readSecrets()
	if err != nil {
		return "", err
	}
	values, ok := secrets[profile]
	if !ok {
		return "", ErrNotFound
	}
	value, ok := values[name]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *FileStore) PutSecret(ctx context.Context, profile string, name string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	secrets, err := s.readSecrets()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if secrets == nil {
		secrets = make(map[string]map[string]string)
	}
	if secrets[profile] == nil {
		secrets[profile] = make(map[string]string)
	}
	secrets[profile][name] = value
	return s.writeSecrets(secrets)
}

func (s *FileStore) DeleteSecret(ctx context.Context, profile string, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	secrets, err := s.readSecrets()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if secrets[profile] != nil {
		delete(secrets[profile], name)
		if len(secrets[profile]) == 0 {
			delete(secrets, profile)
		}
	}
	return s.writeSecrets(secrets)
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

func (s *FileStore) secretsPath() string {
	return s.path + ".secrets.json"
}

func (s *FileStore) readSecrets() (map[string]map[string]string, error) {
	body, err := os.ReadFile(s.secretsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read secret store: %w", err)
	}
	if len(body) == 0 {
		return map[string]map[string]string{}, nil
	}
	var secrets map[string]map[string]string
	if err := json.Unmarshal(body, &secrets); err != nil {
		return nil, fmt.Errorf("decode secret store: %w", err)
	}
	return secrets, nil
}

func (s *FileStore) writeSecrets(secrets map[string]map[string]string) error {
	path := s.secretsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create secret store dir: %w", err)
	}
	body, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("encode secret store: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return fmt.Errorf("write secret store: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace secret store: %w", err)
	}
	return nil
}
