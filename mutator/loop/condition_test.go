package loop

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorLoopCondition(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorLoopCondition,
		"../../testdata/loop/condition.go",
		2,
	)
}
