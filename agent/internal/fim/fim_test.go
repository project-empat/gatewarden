package fim

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	content := []byte("hello-fim")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])

	got, ok := hashFile(path)
	if !ok {
		t.Fatal("hashFile returned not-ok for existing file")
	}
	if got != want {
		t.Errorf("hash = %s, want %s", got, want)
	}
}

func TestHashFileMissing(t *testing.T) {
	if _, ok := hashFile(filepath.Join(t.TempDir(), "nope")); ok {
		t.Error("hashFile should report not-ok for missing file")
	}
}

func TestCheckSkipsMissingAndHashesExisting(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing")
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Point extra paths at one existing file and one missing file via env.
	t.Setenv("GATEWARDEN_FIM_PATHS", existing+":"+filepath.Join(t.TempDir(), "missing"))
	status, err := Check()
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Mode != "periodic" {
		t.Errorf("mode = %q, want periodic", status.Mode)
	}

	var found bool
	for _, f := range status.Files {
		if f.Path == existing {
			found = true
			if f.Hash == "" || len(f.Hash) != 64 {
				t.Errorf("bad hash for %s: %q", existing, f.Hash)
			}
		}
		if f.Path == filepath.Join(t.TempDir(), "missing") {
			t.Error("missing file should be skipped")
		}
	}
	if !found {
		t.Error("existing file was not hashed")
	}
}
