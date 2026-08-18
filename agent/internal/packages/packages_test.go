package packages

import (
	"strings"
	"testing"

	"github.com/gatewarden/agent/pkg/proto"
)

func TestParseDpkgOutput(t *testing.T) {
	input := "openssh-server\t1:9.6p1-3ubuntu13.5\nlibc6\t2.39-0ubuntu8.3\n\nnginx\t1.24\n"
	got := parseDpkgOutput(strings.NewReader(input))
	want := []proto.PackageInfo{
		{Name: "openssh-server", Version: "1:9.6p1-3ubuntu13.5"},
		{Name: "libc6", Version: "2.39-0ubuntu8.3"},
		{Name: "nginx", Version: "1.24"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d packages, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pkg[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseDpkgSkipsBadLines(t *testing.T) {
	input := "\n\n\nfoo\n\tbare-version\nopenssl\t3.0.2\n"
	got := parseDpkgOutput(strings.NewReader(input))
	// Blank lines are dropped; name-only and name+version lines are kept.
	if len(got) != 3 {
		t.Fatalf("got %v, want 3 entries", got)
	}
	if got[0].Name != "foo" || got[0].Version != "" {
		t.Errorf("expected foo with empty version, got %+v", got[0])
	}
	if got[1].Name != "bare-version" {
		t.Errorf("expected bare-version, got %+v", got[1])
	}
	if got[2].Name != "openssl" || got[2].Version != "3.0.2" {
		t.Errorf("expected openssl 3.0.2, got %+v", got[2])
	}
}

func TestParseAptCheckHuman(t *testing.T) {
	out := "5 updates can be applied immediately.\n2 of these updates are security updates.\n"
	pending, security := parseAptCheckHuman(out)
	if pending != 5 || security != 2 {
		t.Errorf("got (%d,%d), want (5,2)", pending, security)
	}

	// No previous line -> pending present, security missing -> 0.
	pending, security = parseAptCheckHuman("7 updates can be applied immediately.\n")
	if pending != 7 || security != 0 {
		t.Errorf("got (%d,%d), want (7,0)", pending, security)
	}

	pending, security = parseAptCheckHuman("nothing here\n")
	if pending != 0 || security != 0 {
		t.Errorf("got (%d,%d), want (0,0)", pending, security)
	}
}
