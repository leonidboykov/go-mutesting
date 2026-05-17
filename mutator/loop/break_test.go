package loop

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorLoopBreak(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorLoopBreak,
		"../../testdata/loop/break.go",
		2,
	)
}
