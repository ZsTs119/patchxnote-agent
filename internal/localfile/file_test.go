package localfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicWriteFileWritesAndRespectsOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "模型 输出.json")
	if err := AtomicWriteFile(path, []byte("first\n"), 0o600, false); err != nil {
		t.Fatalf("write first file: %v", err)
	}
	if err := AtomicWriteFile(path, []byte("second\n"), 0o600, false); err == nil {
		t.Fatal("expected overwrite without force to fail")
	}
	if err := AtomicWriteFile(path, []byte("second\n"), 0o600, true); err != nil {
		t.Fatalf("overwrite file: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(body) != "second\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestAtomicWriteFileRejectsDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := AtomicWriteFile(dir, []byte("body"), 0o600, true); err == nil {
		t.Fatal("expected directory output path to fail")
	}
}

func TestAbsPathResolvesRelativePath(t *testing.T) {
	got, err := AbsPath(filepath.Join("tmp", "模型 输出.txt"))
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}
	if !filepath.IsAbs(got) || !strings.Contains(got, "模型 输出.txt") {
		t.Fatalf("unexpected absolute path: %s", got)
	}
}
