package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathsLinuxUsesXDGWhenPresent(t *testing.T) {
	paths, err := ResolvePaths("linux", PathEnv{
		HomeDir:       "/home/alice",
		XDGConfigHome: "/tmp/config",
		XDGCacheHome:  "/tmp/cache",
	})
	if err != nil {
		t.Fatalf("resolve linux paths: %v", err)
	}

	if paths.ConfigDir != filepath.Join("/tmp/config", "patchxnote-agent") {
		t.Fatalf("unexpected config dir: %s", paths.ConfigDir)
	}
	if paths.ConfigFile != filepath.Join("/tmp/config", "patchxnote-agent", "config.yaml") {
		t.Fatalf("unexpected config file: %s", paths.ConfigFile)
	}
	if paths.CacheDir != filepath.Join("/tmp/cache", "patchxnote-agent") {
		t.Fatalf("unexpected cache dir: %s", paths.CacheDir)
	}
}

func TestResolvePathsDarwin(t *testing.T) {
	paths, err := ResolvePaths("darwin", PathEnv{HomeDir: "/Users/alice"})
	if err != nil {
		t.Fatalf("resolve darwin paths: %v", err)
	}

	if paths.ConfigFile != filepath.Join("/Users/alice", "Library", "Application Support", "PatchXNote Agent", "config.yaml") {
		t.Fatalf("unexpected config file: %s", paths.ConfigFile)
	}
	if paths.CacheDir != filepath.Join("/Users/alice", "Library", "Caches", "PatchXNote Agent") {
		t.Fatalf("unexpected cache dir: %s", paths.CacheDir)
	}
}

func TestResolvePathsWindows(t *testing.T) {
	paths, err := ResolvePaths("windows", PathEnv{
		AppData:      `C:\Users\alice\AppData\Roaming`,
		LocalAppData: `C:\Users\alice\AppData\Local`,
	})
	if err != nil {
		t.Fatalf("resolve windows paths: %v", err)
	}

	if !strings.Contains(paths.ConfigFile, "PatchXNote Agent") || !strings.HasSuffix(paths.ConfigFile, "config.yaml") {
		t.Fatalf("unexpected config file: %s", paths.ConfigFile)
	}
	if !strings.Contains(paths.CacheDir, "PatchXNote Agent") || !strings.HasSuffix(paths.CacheDir, "Cache") {
		t.Fatalf("unexpected cache dir: %s", paths.CacheDir)
	}
}
