// Package remotepath handles POSIX-style paths used by the OpenList API.
// Remote paths always use forward slashes, independent of the local OS.
package remotepath

import (
	"fmt"
	"path"
	"strings"
)

// Clean returns a cleaned absolute remote path.
func Clean(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "/"
	}
	cp := path.Clean(p)
	if cp == "." || cp == "" {
		return "/"
	}
	if !strings.HasPrefix(cp, "/") {
		cp = "/" + cp
	}
	return cp
}

// Join joins remote path elements and cleans the result.
func Join(elem ...string) string {
	return Clean(path.Join(elem...))
}

// Rel returns the path of file relative to root.
func Rel(root, file string) (string, error) {
	root = Clean(root)
	file = Clean(file)

	if root == "/" {
		rel := strings.TrimPrefix(file, "/")
		if rel == "" {
			return "", fmt.Errorf("invalid relative path for %q", file)
		}
		return rel, nil
	}

	if file == root {
		return path.Base(file), nil
	}

	prefix := root + "/"
	if !strings.HasPrefix(file, prefix) {
		return "", fmt.Errorf("source path %q is not under root %q", file, root)
	}

	rel := strings.TrimPrefix(file, prefix)
	if rel == "" {
		return "", fmt.Errorf("invalid relative path for %q", file)
	}
	return rel, nil
}
