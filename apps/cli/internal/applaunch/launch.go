package applaunch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/richinosan/akari-video/apps/cli/internal/reporoot"
)

const (
	bundledAppDir = "lib/app"
	appBundleName = "AKARI Video.app"
)

// Run starts the AKARI Video desktop app. Optional args are workspace paths.
func Run(workspaceArgs []string) error {
	if binary, appRoot, ok := resolveBundled(); ok {
		return runCommand(binary, appRoot, workspaceArgs)
	}
	if binary, shellRoot, ok := resolveDev(); ok {
		args := append([]string{shellRoot}, workspaceArgs...)
		return runCommand(binary, shellRoot, args)
	}
	return fmt.Errorf("AKARI Video app not found (expected %s next to akari binary or monorepo apps/shell)", bundledAppDir)
}

func runCommand(binary, dir string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func resolveBundled() (binary string, appRoot string, ok bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", false
	}
	appRoot = filepath.Join(filepath.Dir(exe), bundledAppDir)
	if _, err := os.Stat(appRoot); err != nil {
		return "", "", false
	}
	binary, ok = findAppBinary(appRoot)
	if !ok {
		return "", "", false
	}
	return binary, appRoot, true
}

func resolveDev() (binary string, shellRoot string, ok bool) {
	repo, err := reporoot.FromEnvOrResolve("")
	if err != nil {
		return "", "", false
	}
	shellRoot = filepath.Join(repo, "apps", "shell")
	mainJS := filepath.Join(shellRoot, "lib", "backend", "electron-main.js")
	if _, err := os.Stat(mainJS); err != nil {
		return "", "", false
	}

	switch runtime.GOOS {
	case "linux", "windows":
		binary = filepath.Join(shellRoot, "node_modules", "electron", "dist", electronExecutableName())
	case "darwin":
		binary = filepath.Join(shellRoot, "node_modules", "electron", "dist", "Electron.app", "Contents", "MacOS", "Electron")
	default:
		return "", "", false
	}
	if _, err := os.Stat(binary); err != nil {
		return "", "", false
	}
	return binary, shellRoot, true
}

func electronExecutableName() string {
	if runtime.GOOS == "windows" {
		return "electron.exe"
	}
	return "electron"
}

func findAppBinary(appRoot string) (string, bool) {
	if runtime.GOOS == "darwin" {
		binary := filepath.Join(appRoot, appBundleName, "Contents", "MacOS", "AKARI Video")
		if fileExists(binary) {
			return binary, true
		}
	}

	candidates := []string{"akari-video", "AKARI Video", "@akari-videoshell"}
	if runtime.GOOS == "windows" {
		candidates = []string{"akari-video.exe", "AKARI Video.exe"}
	}
	for _, name := range candidates {
		binary := filepath.Join(appRoot, name)
		if fileExists(binary) {
			return binary, true
		}
	}

	entries, err := os.ReadDir(appRoot)
	if err != nil {
		return "", false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if runtime.GOOS == "windows" {
			if !strings.HasSuffix(strings.ToLower(name), ".exe") {
				continue
			}
		} else if strings.Contains(name, ".") {
			continue
		}
		binary := filepath.Join(appRoot, name)
		if isExecutable(binary) {
			return binary, true
		}
	}
	return "", false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}
