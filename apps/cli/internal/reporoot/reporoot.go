package reporoot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var marker = filepath.Join("packages", "render-cut", "bin", "render-cut.mjs")

// Resolve walks upward from start (default cwd) until it finds the monorepo root.
func Resolve(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for dir := abs; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return "", fmt.Errorf("akari-video repo root not found from %s (expected %s)", start, marker)
}

// MustResolve panics when Resolve fails. Tests only.
func MustResolve(start string) string {
	root, err := Resolve(start)
	if err != nil {
		panic(err)
	}
	return root
}

// FromEnvOrResolve uses AKARI_VIDEO_ROOT when set, otherwise Resolve.
func FromEnvOrResolve(start string) (string, error) {
	if env := os.Getenv("AKARI_VIDEO_ROOT"); env != "" {
		abs, err := filepath.Abs(env)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(filepath.Join(abs, marker)); err != nil {
			return "", fmt.Errorf("AKARI_VIDEO_ROOT=%s does not contain %s: %w", abs, marker, err)
		}
		return abs, nil
	}
	return Resolve(start)
}

var ErrNotFound = errors.New("repo root not found")
