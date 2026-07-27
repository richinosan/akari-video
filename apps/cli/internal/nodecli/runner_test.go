package nodecli_test

import (
	"testing"

	akariv1 "github.com/richinosan/akari-video/apps/cli/gen/akari/v1"
	"github.com/richinosan/akari-video/apps/cli/internal/nodecli"
)

func TestExitClassMapsRenderCutCodes(t *testing.T) {
	tests := []struct {
		code int
		want akariv1.ExitClass
	}{
		{0, akariv1.ExitClass_EXIT_CLASS_OK},
		{1, akariv1.ExitClass_EXIT_CLASS_REFUSAL},
		{2, akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR},
		{9, akariv1.ExitClass_EXIT_CLASS_EXECUTION_ERROR},
	}
	for _, tc := range tests {
		if got := nodecli.ExitClass(tc.code); got != tc.want {
			t.Fatalf("ExitClass(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}
