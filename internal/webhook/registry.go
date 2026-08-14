package webhook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"gopkg.in/yaml.v3"
)

var ErrTargetNotFound = errors.New("webhook target not found")

type Registry struct {
	cfg     config.Config
	profile string
	targets map[string]Target
	now     func() time.Time
}

func NewRegistry(cfg config.Config) *Registry {
	profile := cfg.Profile
	if profile == "" {
		profile = "default"
	}
	registry := &Registry{
		cfg:     cfg,
		profile: profile,
		targets: make(map[string]Target),
		now:     time.Now,
	}
	if cfg.Paths.ConfigFile != "" {
		if targets := readTargetsFromConfigFile(cfg.Paths.ConfigFile, profile); len(targets) > 0 {
			registry.targets = targets
			return registry
		}
	}
	if cfg.Webhooks.Profiles != nil {
		profileConfig := cfg.Webhooks.Profiles[profile]
		for key, targetConfig := range profileConfig.Targets {
			target, err := targetFromConfig(key, targetConfig)
			if err == nil {
				registry.targets[target.Alias] = target
			}
		}
	}
	return registry
}

func readTargetsFromConfigFile(path string, profile string) map[string]Target {
	root, err := readConfigMap(path)
	if err != nil {
		return nil
	}
	webhooks, ok := root["webhooks"].(map[string]any)
	if !ok {
		return nil
	}
	profiles, ok := webhooks["profiles"].(map[string]any)
	if !ok {
		return nil
	}
	profileConfig, ok := profiles[profile].(map[string]any)
	if !ok {
		return nil
	}
	targetConfigs, ok := profileConfig["targets"].(map[string]any)
	if !ok {
		return nil
	}
	targets := make(map[string]Target, len(targetConfigs))
	for key, raw := range targetConfigs {
		var targetConfig config.WebhookTargetConfig
		body, err := yaml.Marshal(raw)
		if err != nil {
			continue
		}
		if err := yaml.Unmarshal(body, &targetConfig); err != nil {
			continue
		}
		target, err := targetFromConfig(key, targetConfig)
		if err != nil {
			continue
		}
		targets[target.Alias] = target
	}
	return targets
}

func (r *Registry) List() []Target {
	items := make([]Target, 0, len(r.targets))
	for _, target := range r.targets {
		items = append(items, target)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Alias < items[j].Alias
	})
	return items
}

func (r *Registry) Get(alias string) (Target, error) {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return Target{}, err
	}
	target, ok := r.targets[normalized]
	if !ok {
		return Target{}, fmt.Errorf("%w: %s", ErrTargetNotFound, normalized)
	}
	return target, nil
}

func (r *Registry) Set(alias string, targetType TargetType, maskedURL string, template string) (Target, error) {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return Target{}, err
	}
	if _, err := ValidateType(string(targetType)); err != nil {
		return Target{}, err
	}
	now := r.now().UTC()
	target, exists := r.targets[normalized]
	if !exists {
		target = Target{
			Alias:     normalized,
			Enabled:   true,
			CreatedAt: now,
		}
	}
	target.Type = targetType
	target.MaskedURL = maskedURL
	target.Template = template
	target.UpdatedAt = now
	r.targets[normalized] = target
	return target, r.save()
}

func (r *Registry) SetEnabled(alias string, enabled bool) (Target, error) {
	target, err := r.Get(alias)
	if err != nil {
		return Target{}, err
	}
	target.Enabled = enabled
	target.UpdatedAt = r.now().UTC()
	r.targets[target.Alias] = target
	return target, r.save()
}

func (r *Registry) Remove(alias string) (Target, error) {
	target, err := r.Get(alias)
	if err != nil {
		return Target{}, err
	}
	delete(r.targets, target.Alias)
	return target, r.save()
}

func (r *Registry) save() error {
	if r.cfg.Paths.ConfigFile == "" {
		return fmt.Errorf("config file path is required")
	}

	root, err := readConfigMap(r.cfg.Paths.ConfigFile)
	if err != nil {
		return err
	}
	profiles := ensureMap(root, "webhooks", "profiles")
	profileMap := ensureMap(profiles, r.profile)
	targetsMap := make(map[string]any, len(r.targets))
	for _, target := range r.targets {
		targetsMap[target.Alias] = targetToConfig(target)
	}
	profileMap["targets"] = targetsMap

	body, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(r.cfg.Paths.ConfigFile), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	tmp := r.cfg.Paths.ConfigFile + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, r.cfg.Paths.ConfigFile); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func readConfigMap(path string) (map[string]any, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	if len(body) == 0 {
		return map[string]any{}, nil
	}
	root := make(map[string]any)
	if err := yaml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if root == nil {
		root = make(map[string]any)
	}
	return root, nil
}

func ensureMap(root map[string]any, keys ...string) map[string]any {
	current := root
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			next = make(map[string]any)
			current[key] = next
		}
		current = next
	}
	return current
}

func targetFromConfig(key string, cfg config.WebhookTargetConfig) (Target, error) {
	alias := cfg.Alias
	if alias == "" {
		alias = key
	}
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return Target{}, err
	}
	targetType, err := ValidateType(cfg.Type)
	if err != nil {
		return Target{}, err
	}
	createdAt, _ := time.Parse(time.RFC3339Nano, cfg.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339Nano, cfg.UpdatedAt)
	return Target{
		Alias:     normalized,
		Type:      targetType,
		Enabled:   cfg.Enabled,
		MaskedURL: cfg.MaskedURL,
		Template:  cfg.Template,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func targetToConfig(target Target) config.WebhookTargetConfig {
	return config.WebhookTargetConfig{
		Alias:     target.Alias,
		Type:      string(target.Type),
		Enabled:   target.Enabled,
		MaskedURL: target.MaskedURL,
		Template:  target.Template,
		CreatedAt: target.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt: target.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
