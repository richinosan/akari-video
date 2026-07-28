package applaunch

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindAppBinaryBundledLayout(t *testing.T) {
	appRoot := t.TempDir()
	name := "akari-video"
	if runtime.GOOS == "windows" {
		name = "akari-video.exe"
	}
	binary := filepath.Join(appRoot, name)
	if err := os.WriteFile(binary, []byte{}, 0o755); err != nil {
		t.Fatal(err)
	}

	got, ok := findAppBinary(appRoot)
	if !ok {
		t.Fatal("expected bundled app binary to be found")
	}
	if got != binary {
		t.Fatalf("got %q want %q", got, binary)
	}
}
