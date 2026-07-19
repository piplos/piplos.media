package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// SafeJoin resolves rel inside root and returns an absolute path.
// The leading-slash Clean trick neutralizes ".." segments, so the result
// cannot escape root; empty relative paths are rejected.
func SafeJoin(root, rel string) (string, error) {
	clean := path.Clean("/" + strings.ReplaceAll(rel, "\\", "/"))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." {
		return "", fmt.Errorf("invalid relative path %q", rel)
	}
	abs := filepath.Join(root, filepath.FromSlash(clean))
	if !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes root", rel)
	}
	return abs, nil
}
