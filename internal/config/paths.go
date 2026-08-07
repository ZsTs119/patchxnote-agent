package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Paths struct {
	ConfigDir  string `json:"config_dir"`
	ConfigFile string `json:"config_file"`
	CacheDir   string `json:"cache_dir"`
}

type PathEnv struct {
	HomeDir       string
	XDGConfigHome string
	XDGCacheHome  string
	AppData       string
	LocalAppData  string
}

func DefaultPathEnv() PathEnv {
	home, _ := os.UserHomeDir()
	return PathEnv{
		HomeDir:       home,
		XDGConfigHome: os.Getenv("XDG_CONFIG_HOME"),
		XDGCacheHome:  os.Getenv("XDG_CACHE_HOME"),
		AppData:       os.Getenv("APPDATA"),
		LocalAppData:  os.Getenv("LOCALAPPDATA"),
	}
}

func ResolvePaths(targetOS string, env PathEnv) (Paths, error) {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if env == (PathEnv{}) {
		env = DefaultPathEnv()
	}

	switch targetOS {
	case "darwin":
		if env.HomeDir == "" {
			return Paths{}, fmt.Errorf("resolve config paths: home directory is required for darwin")
		}
		configDir := filepath.Join(env.HomeDir, "Library", "Application Support", "PatchXNote Agent")
		cacheDir := filepath.Join(env.HomeDir, "Library", "Caches", "PatchXNote Agent")
		return Paths{
			ConfigDir:  configDir,
			ConfigFile: filepath.Join(configDir, "config.yaml"),
			CacheDir:   cacheDir,
		}, nil
	case "windows":
		appData := env.AppData
		if appData == "" && env.HomeDir != "" {
			appData = filepath.Join(env.HomeDir, "AppData", "Roaming")
		}
		localAppData := env.LocalAppData
		if localAppData == "" && env.HomeDir != "" {
			localAppData = filepath.Join(env.HomeDir, "AppData", "Local")
		}
		if appData == "" || localAppData == "" {
			return Paths{}, fmt.Errorf("resolve config paths: APPDATA and LOCALAPPDATA are required for windows")
		}
		configDir := filepath.Join(appData, "PatchXNote Agent")
		cacheDir := filepath.Join(localAppData, "PatchXNote Agent", "Cache")
		return Paths{
			ConfigDir:  configDir,
			ConfigFile: filepath.Join(configDir, "config.yaml"),
			CacheDir:   cacheDir,
		}, nil
	case "linux":
		configBase := env.XDGConfigHome
		if configBase == "" {
			if env.HomeDir == "" {
				return Paths{}, fmt.Errorf("resolve config paths: home directory is required for linux")
			}
			configBase = filepath.Join(env.HomeDir, ".config")
		}
		cacheBase := env.XDGCacheHome
		if cacheBase == "" {
			if env.HomeDir == "" {
				return Paths{}, fmt.Errorf("resolve cache paths: home directory is required for linux")
			}
			cacheBase = filepath.Join(env.HomeDir, ".cache")
		}
		configDir := filepath.Join(configBase, "patchxnote-agent")
		cacheDir := filepath.Join(cacheBase, "patchxnote-agent")
		return Paths{
			ConfigDir:  configDir,
			ConfigFile: filepath.Join(configDir, "config.yaml"),
			CacheDir:   cacheDir,
		}, nil
	default:
		return Paths{}, fmt.Errorf("resolve config paths: unsupported os %q", targetOS)
	}
}
