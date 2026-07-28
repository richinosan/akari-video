package nodecli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
)

// ExitClass maps a child process exit code to the shared protobuf enum.
func ExitClass(code int) akariv1.ExitClass {
	switch code {
	case 0:
		return akariv1.ExitClass_EXIT_CLASS_OK
	case 1:
		return akariv1.ExitClass_EXIT_CLASS_REFUSAL
	case 2:
		return akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR
	default:
		return akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR
	}
}

type Runner struct {
	RepoRoot string
	NodeBin  string
}

func NewRunner(repoRoot, nodeBin string) *Runner {
	if nodeBin == "" {
		nodeBin = "node"
	}
	return &Runner{RepoRoot: repoRoot, NodeBin: nodeBin}
}

func (r *Runner) RenderCut(ctx context.Context, projectRoot string, planOnly bool, outputPath string, force bool) (int, string, error) {
	script := filepath.Join(r.RepoRoot, "packages", "render-cut", "bin", "render-cut.mjs")
	args := []string{script, projectRoot}
	if planOnly {
		args = append(args, "--plan-only")
	}
	if outputPath != "" {
		args = append(args, "--out", outputPath)
	}
	if force {
		args = append(args, "--force")
	}
	return r.run(ctx, args)
}

func (r *Runner) EditLint(ctx context.Context, projectRoot string, includeMedia bool) (int, string, error) {
	script := filepath.Join(r.RepoRoot, "packages", "edit-lint", "bin", "edit-lint.mjs")
	args := []string{script, projectRoot}
	if includeMedia {
		args = append(args, "--media")
	}
	return r.run(ctx, args)
}

func (r *Runner) run(ctx context.Context, args []string) (int, string, error) {
	cmd := exec.CommandContext(ctx, r.NodeBin, args...)
	cmd.Dir = r.RepoRoot
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	err := cmd.Run()
	if err == nil {
		return 0, strings.TrimSpace(combined.String()), nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), strings.TrimSpace(combined.String()), nil
	}
	return 2, strings.TrimSpace(combined.String()), fmt.Errorf("spawn %s: %w", r.NodeBin, err)
}

// NodeFromMise returns node from PATH; callers may set AKARI_NODE to override.
func NodeFromEnv() string {
	if v := os.Getenv("AKARI_NODE"); v != "" {
		return v
	}
	return "node"
}
