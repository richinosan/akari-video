package nodebundle_test

import (
	"os"
	"testing"

	"github.com/richinosan/akari-video/apps/cli/internal/nodebundle"
)

func TestRuntimeRootFromMonorepo(t *testing.T) {
	root := mustResolveForTest(t)
	if _, err := os.Stat(nodebundle.OrchestratorScript(root)); err != nil {
		t.Fatalf("orchestrator script missing: %v", err)
	}
}

func mustResolveForTest(t *testing.T) string {
	t.Helper()
	root, err := nodebundle.RuntimeRoot()
	if err != nil {
		t.Fatalf("RuntimeRoot: %v", err)
	}
	return root
}
