package utils

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeJoin(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "srv", "store")

	valid := map[string]string{
		"a.txt":        filepath.Join(root, "a.txt"),
		"a/b/c.tar.gz": filepath.Join(root, "a", "b", "c.tar.gz"),
		"./a/../b":     filepath.Join(root, "b"),
		"a\\b":         filepath.Join(root, "a", "b"),
	}
	for rel, want := range valid {
		got, err := SafeJoin(root, rel)
		if err != nil {
			t.Errorf("SafeJoin(%q): unexpected error %v", rel, err)
			continue
		}
		if got != want {
			t.Errorf("SafeJoin(%q) = %q, want %q", rel, got, want)
		}
	}

	// «..» нейтрализуется, а не выходит за пределы root.
	for _, rel := range []string{"../evil", "a/../../evil", "..\\evil"} {
		got, err := SafeJoin(root, rel)
		if err != nil {
			t.Errorf("SafeJoin(%q): unexpected error %v", rel, err)
			continue
		}
		if !strings.HasPrefix(got, root+string(filepath.Separator)) {
			t.Errorf("SafeJoin(%q) = %q escaped root", rel, got)
		}
	}

	for _, rel := range []string{"", ".", "/", "//", "..", "a/.."} {
		if _, err := SafeJoin(root, rel); err == nil {
			t.Errorf("SafeJoin(%q): expected error", rel)
		}
	}
}
