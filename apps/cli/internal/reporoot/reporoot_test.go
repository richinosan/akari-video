package reporoot_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/richinosan/akari-video/apps/cli/internal/reporoot"
)

func TestResolveFindsRepoRoot(t *testing.T) {
	root := reporoot.MustResolve("")
	marker := filepath.Join(root, "packages", "render-cut", "bin", "render-cut.mjs")
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("marker missing: %v", err)
	}
}

func TestFromEnvOrResolveRejectsInvalidEnv(t *testing.T) {
	t.Setenv("AKARI_VIDEO_ROOT", t.TempDir())
	_, err := reporoot.FromEnvOrResolve("")
	if err == nil {
		t.Fatal("expected error for invalid AKARI_VIDEO_ROOT")
	}
}
