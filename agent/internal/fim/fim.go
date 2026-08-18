package fim

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/gatewarden/agent/pkg/proto"
)

// defaultPaths are critical files monitored for tampering. Directories are
// scanned for regular files recursively (bounded).
var defaultPaths = []string{
	"/etc/passwd",
	"/etc/shadow",
	"/etc/ssh/sshd_config",
	"/etc/crontab",
	"/etc/hosts",
	"/etc/hosts.allow",
	"/etc/hosts.deny",
	"/etc/sudoers",
	"/etc/systemd/system",
	"/etc/ufw",
}

const maxHashBytes = 64 * 1024 * 1024 // cap per-file hashing at 64 MB

// Check hashes the configured set of critical files and returns a FIM status.
// Mode is "periodic"; the premium realtime (fanotify) mode is added in the
// enterprise agent build.
func Check() (*proto.FIMStatus, error) {
	var files []proto.FIMFile

	seen := map[string]bool{}
	for _, path := range append(append([]string{}, defaultPaths...), extraPaths()...) {
		if seen[path] {
			continue
		}
		seen[path] = true

		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			for _, child := range hashDir(path) {
				files = append(files, child)
			}
			continue
		}
		if h, ok := hashFile(path); ok {
			files = append(files, proto.FIMFile{Path: path, Hash: h})
		}
	}

	if len(files) == 0 {
		return nil, nil
	}
	return &proto.FIMStatus{Mode: "periodic", Files: files}, nil
}

// extraPaths extends the monitored set via GATEWARDEN_FIM_PATHS (colon or
// comma separated), e.g. /etc/myapp/app.conf.
func extraPaths() []string {
	raw := os.Getenv("GATEWARDEN_FIM_PATHS")
	if raw == "" {
		return nil
	}
	return strings.FieldsFunc(raw, func(r rune) bool { return r == ':' || r == ',' })
}

func hashFile(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	h := sha256.New()
	buf := make([]byte, 32*1024)
	remaining := int64(maxHashBytes)
	for {
		if remaining <= 0 {
			break
		}
		n, err := f.Read(buf[:minInt64(int64(len(buf)), remaining)])
		if n > 0 {
			h.Write(buf[:n])
			remaining -= int64(n)
		}
		if err != nil {
			break
		}
	}
	return hex.EncodeToString(h.Sum(nil)), true
}

func hashDir(dir string) []proto.FIMFile {
	var out []proto.FIMFile
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if h, ok := hashFile(path); ok {
			out = append(out, proto.FIMFile{Path: path, Hash: h})
		}
		return nil
	})
	return out
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
