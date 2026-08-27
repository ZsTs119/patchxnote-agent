package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
)

type ConfigTarget struct {
	Path         string
	ExpectedRoot string
	Warning      string
}

func ResolveClientConfigTarget(clientID string, targetOS string, env config.PathEnv) (ConfigTarget, error) {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if env == (config.PathEnv{}) {
		env = config.DefaultPathEnv()
	}
	switch targetOS {
	case "windows":
		return windowsConfigTarget(clientID, env)
	case "darwin":
		return darwinConfigTarget(clientID, env)
	case "linux":
		return linuxConfigTarget(clientID, env)
	default:
		return ConfigTarget{}, fmt.Errorf("unsupported setup runtime %q", targetOS)
	}
}

func ensureHome(env config.PathEnv, targetOS string) (string, error) {
	if env.HomeDir != "" {
		return env.HomeDir, nil
	}
	if targetOS == "windows" {
		if env.AppData != "" {
			return parentDir(parentDir(env.AppData)), nil
		}
		if env.LocalAppData != "" {
			return parentDir(parentDir(env.LocalAppData)), nil
		}
	}
	return "", fmt.Errorf("home directory is required for %s client setup", targetOS)
}

func windowsAppData(env config.PathEnv) (string, error) {
	if env.AppData != "" {
		return env.AppData, nil
	}
	home, err := ensureHome(env, "windows")
	if err != nil {
		return "", err
	}
	return joinForOS("windows", home, "AppData", "Roaming"), nil
}

func linuxConfigHome(env config.PathEnv) (string, error) {
	if env.XDGConfigHome != "" {
		return env.XDGConfigHome, nil
	}
	home, err := ensureHome(env, "linux")
	if err != nil {
		return "", err
	}
	return joinForOS("linux", home, ".config"), nil
}

func joinForOS(targetOS string, parts ...string) string {
	sep := "/"
	if targetOS == "windows" {
		sep = "\\"
	}
	var cleaned []string
	for index, part := range parts {
		if part == "" {
			continue
		}
		if index == 0 {
			cleaned = append(cleaned, strings.TrimRight(part, `\/`))
			continue
		}
		cleaned = append(cleaned, strings.Trim(part, `\/`))
	}
	if len(cleaned) == 0 {
		return ""
	}
	return strings.Join(cleaned, sep)
}

func normalizePathForCompare(targetOS string, value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimRight(value, "/")
	if targetOS == "windows" {
		value = strings.ToLower(value)
	}
	return value
}

func isPathUnder(targetOS string, child string, root string) bool {
	child = normalizePathForCompare(targetOS, child)
	root = normalizePathForCompare(targetOS, root)
	if child == root {
		return true
	}
	return strings.HasPrefix(child, root+"/")
}

func safeAbs(pathValue string) string {
	abs, err := filepath.Abs(pathValue)
	if err != nil {
		return pathValue
	}
	return abs
}

func parentDir(value string) string {
	trimmed := strings.TrimRight(value, `\/`)
	index := strings.LastIndexAny(trimmed, `\/`)
	if index <= 0 {
		return ""
	}
	return trimmed[:index]
}

func runningTargetOS() string {
	return runtime.GOOS
}

func pathExists(pathValue string) bool {
	_, err := os.Stat(pathValue)
	return err == nil
}
