package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Profile  string        `mapstructure:"profile" json:"profile"`
	Output   string        `mapstructure:"output" json:"output"`
	Server   ServerConfig  `mapstructure:"server" json:"server"`
	Auth     AuthConfig    `mapstructure:"auth" json:"auth"`
	Webhooks WebhookConfig `mapstructure:"webhooks" json:"webhooks"`
	Paths    Paths         `mapstructure:"-" json:"paths"`
}

type ServerConfig struct {
	BaseURL string `mapstructure:"base_url" json:"base_url"`
}

const DefaultServerBaseURL = "https://ws-lab.patch-x.cn/patchnote-test-api"

type AuthConfig struct {
	InsecureFileKeychain bool `mapstructure:"insecure_file_keychain" json:"insecure_file_keychain"`
}

type WebhookConfig struct {
	Profiles map[string]WebhookProfileConfig `mapstructure:"profiles" json:"profiles" yaml:"profiles"`
}

type WebhookProfileConfig struct {
	Targets map[string]WebhookTargetConfig `mapstructure:"targets" json:"targets" yaml:"targets"`
}

type WebhookTargetConfig struct {
	Alias     string `mapstructure:"alias" json:"alias" yaml:"alias"`
	Type      string `mapstructure:"type" json:"type" yaml:"type"`
	Enabled   bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	MaskedURL string `mapstructure:"masked_url" json:"masked_url" yaml:"masked_url"`
	Template  string `mapstructure:"template,omitempty" json:"template,omitempty" yaml:"template,omitempty"`
	CreatedAt string `mapstructure:"created_at" json:"created_at" yaml:"created_at"`
	UpdatedAt string `mapstructure:"updated_at" json:"updated_at" yaml:"updated_at"`
}

type LoadOptions struct {
	ConfigFile string
	GOOS       string
	PathEnv    PathEnv
}

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("PATCHXNOTE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetDefault("profile", "default")
	v.SetDefault("output", "plain")
	v.SetDefault("server.base_url", DefaultServerBaseURL)
	v.SetDefault("auth.insecure_file_keychain", false)
	mustBindEnv(v, "config", "PATCHXNOTE_CONFIG", "PATCHNOTE_CONFIG")
	mustBindEnv(v, "profile", "PATCHXNOTE_PROFILE", "PATCHNOTE_PROFILE")
	mustBindEnv(v, "output", "PATCHXNOTE_OUTPUT", "PATCHNOTE_OUTPUT")
	mustBindEnv(v, "server.base_url", "PATCHXNOTE_SERVER_BASE_URL", "PATCHNOTE_SERVER_BASE_URL")
	mustBindEnv(v, "auth.insecure_file_keychain", "PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN", "PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN")
	v.AutomaticEnv()
	return v
}

func mustBindEnv(v *viper.Viper, key string, names ...string) {
	if err := v.BindEnv(append([]string{key}, names...)...); err != nil {
		panic(err)
	}
}

func Load(v *viper.Viper, opts LoadOptions) (Config, error) {
	if v == nil {
		v = NewViper()
	}

	paths, err := ResolvePaths(opts.GOOS, opts.PathEnv)
	if err != nil {
		return Config{}, err
	}

	configFile := opts.ConfigFile
	if configFile == "" {
		configFile = v.GetString("config")
	}
	explicitConfigFile := configFile != ""
	if configFile == "" {
		configFile = paths.ConfigFile
	}

	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		defaultConfigMissing := !explicitConfigFile && (errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist))
		if !defaultConfigMissing {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if cfg.Profile == "" {
		cfg.Profile = "default"
	}
	if cfg.Output == "" {
		cfg.Output = "plain"
	}
	cfg.Paths = paths
	cfg.Paths.ConfigFile = configFile

	return cfg, nil
}
