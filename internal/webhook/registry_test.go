package webhook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
)

func TestRegistryWritesAndLoadsTargets(t *testing.T) {
	cfg := registryTestConfig(t, "default")
	fixedNow := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	registry := NewRegistry(cfg)
	registry.now = func() time.Time { return fixedNow }

	target, err := registry.Set(" 产品群 飞书 ", TargetTypeFeishu, "https://open.feishu.cn/.../c92b...64cf", "default")
	if err != nil {
		t.Fatalf("set target: %v", err)
	}
	if target.Alias != "产品群 飞书" || !target.Enabled {
		t.Fatalf("unexpected target: %+v", target)
	}

	loaded, err := config.Load(config.NewViper(), config.LoadOptions{
		ConfigFile: cfg.Paths.ConfigFile,
		GOOS:       "linux",
		PathEnv:    config.PathEnv{HomeDir: filepath.Dir(cfg.Paths.ConfigFile)},
	})
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}
	reloaded := NewRegistry(loaded)
	got, err := reloaded.Get("产品群 飞书")
	if err != nil {
		t.Fatalf("get reloaded: %v", err)
	}
	if got.Type != TargetTypeFeishu || got.MaskedURL == "" || got.CreatedAt.IsZero() {
		t.Fatalf("unexpected reloaded target: %+v", got)
	}
}

func TestRegistryLoadsAliasWithDotsFromConfigFile(t *testing.T) {
	cfg := registryTestConfig(t, "default")
	alias := "0.2.5发布验收 飞书"

	if _, err := NewRegistry(cfg).Set(alias, TargetTypeFeishu, "https://open.feishu.cn/open-apis/bot/v2/hook/c92b...64cf", "default"); err != nil {
		t.Fatalf("set target: %v", err)
	}
	loaded, err := config.Load(config.NewViper(), config.LoadOptions{
		ConfigFile: cfg.Paths.ConfigFile,
		GOOS:       "linux",
		PathEnv:    config.PathEnv{HomeDir: filepath.Dir(cfg.Paths.ConfigFile)},
	})
	if err != nil {
		t.Fatalf("reload config: %v", err)
	}

	reloaded := NewRegistry(loaded)
	got, err := reloaded.Get(alias)
	if err != nil {
		t.Fatalf("get alias with dots: %v", err)
	}
	if got.Alias != alias || got.Type != TargetTypeFeishu {
		t.Fatalf("unexpected target: %+v", got)
	}
}

func TestRegistryUpdatesExistingWithoutDuplicate(t *testing.T) {
	cfg := registryTestConfig(t, "default")
	registry := NewRegistry(cfg)
	registry.now = func() time.Time { return time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC) }

	if _, err := registry.Set("内部网关", TargetTypeGeneric, "http://127.0.0.1/hook", "default"); err != nil {
		t.Fatalf("set: %v", err)
	}
	registry.now = func() time.Time { return time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC) }
	if _, err := registry.Set("内部网关", TargetTypeDingTalk, "https://oapi.dingtalk.com/robot/send?access_token=***", "default"); err != nil {
		t.Fatalf("update: %v", err)
	}

	items := registry.List()
	if len(items) != 1 || items[0].Type != TargetTypeDingTalk {
		t.Fatalf("expected one updated target, got %+v", items)
	}
	if !items[0].UpdatedAt.After(items[0].CreatedAt) {
		t.Fatalf("expected updated timestamp after created timestamp: %+v", items[0])
	}
}

func TestRegistryEnableDisableRemoveAndPreserveConfig(t *testing.T) {
	cfg := registryTestConfig(t, "default")
	initial := "profile: default\nserver:\n  base_url: https://file.example\n"
	if err := os.WriteFile(cfg.Paths.ConfigFile, []byte(initial), 0o600); err != nil {
		t.Fatalf("write initial config: %v", err)
	}
	registry := NewRegistry(cfg)

	if _, err := registry.Set("产品群", TargetTypeFeishu, "masked", "default"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, err := registry.SetEnabled("产品群", false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	disabled, err := registry.Get("产品群")
	if err != nil {
		t.Fatalf("get disabled: %v", err)
	}
	if disabled.Enabled {
		t.Fatal("expected disabled target")
	}
	if _, err := registry.SetEnabled("产品群", true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if _, err := registry.Remove("产品群"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := registry.Get("产品群"); err == nil {
		t.Fatal("expected removed target to be missing")
	}

	body, err := os.ReadFile(cfg.Paths.ConfigFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(body), "base_url: https://file.example") {
		t.Fatalf("config did not preserve unrelated fields:\n%s", string(body))
	}
}

func TestRegistryIsolatesProfiles(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	cfgA := registryConfigWithPath("profile-a", configFile)
	cfgB := registryConfigWithPath("profile-b", configFile)

	if _, err := NewRegistry(cfgA).Set("同名目标", TargetTypeFeishu, "masked-a", "default"); err != nil {
		t.Fatalf("set profile a: %v", err)
	}
	if _, err := NewRegistry(cfgB).Set("同名目标", TargetTypeGeneric, "masked-b", "default"); err != nil {
		t.Fatalf("set profile b: %v", err)
	}

	loadedA, err := config.Load(config.NewViper(), config.LoadOptions{
		ConfigFile: configFile,
		GOOS:       "linux",
		PathEnv:    config.PathEnv{HomeDir: dir},
	})
	if err != nil {
		t.Fatalf("load profile a default: %v", err)
	}
	loadedA.Profile = "profile-a"
	targetA, err := NewRegistry(loadedA).Get("同名目标")
	if err != nil {
		t.Fatalf("get profile a: %v", err)
	}
	loadedA.Profile = "profile-b"
	targetB, err := NewRegistry(loadedA).Get("同名目标")
	if err != nil {
		t.Fatalf("get profile b: %v", err)
	}
	if targetA.Type != TargetTypeFeishu || targetB.Type != TargetTypeGeneric {
		t.Fatalf("profiles not isolated: a=%+v b=%+v", targetA, targetB)
	}
}

func registryTestConfig(t *testing.T, profile string) config.Config {
	t.Helper()
	return registryConfigWithPath(profile, filepath.Join(t.TempDir(), "config.yaml"))
}

func registryConfigWithPath(profile string, configFile string) config.Config {
	return config.Config{
		Profile: profile,
		Output:  "plain",
		Paths: config.Paths{
			ConfigFile: configFile,
			ConfigDir:  filepath.Dir(configFile),
		},
	}
}
