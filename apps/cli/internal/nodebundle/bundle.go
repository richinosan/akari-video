package nodebundle

import (
	"os"
	"path/filepath"

	"github.com/richinosan/akari-video/apps/cli/internal/reporoot"
)

const orchestratorRel = "packages/orchestrator/bin/akari-orchestrator.mjs"

func hasOrchestrator(root string) bool {
	_, err := os.Stat(filepath.Join(root, orchestratorRel))
	return err == nil
}

// RuntimeRoot returns the directory that contains packages/orchestrator (bundle or monorepo).
func RuntimeRoot() (string, error) {
	if exe, err := os.Executable(); err == nil {
		adjacent := filepath.Join(filepath.Dir(exe), "akari-bundle")
		if hasOrchestrator(adjacent) {
			return adjacent, nil
		}
	}

	repo, err := reporoot.FromEnvOrResolve("")
	if err == nil && hasOrchestrator(repo) {
		return repo, nil
	}

	return "", errRuntimeNotFound()
}

func OrchestratorScript(runtimeRoot string) string {
	return filepath.Join(runtimeRoot, orchestratorRel)
}

// NodeBinary returns the node executable (AKARI_NODE or PATH).
func NodeBinary() string {
	if v := os.Getenv("AKARI_NODE"); v != "" {
		return v
	}
	return "node"
}

type runtimeNotFoundError struct{}

func (runtimeNotFoundError) Error() string {
	return "orchestrator runtime not found (expected bin/akari-bundle or monorepo packages/orchestrator)"
}

func errRuntimeNotFound() error {
	return runtimeNotFoundError{}
}
