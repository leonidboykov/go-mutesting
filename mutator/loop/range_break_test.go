package loop

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorLoopRangeBreak(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorLoopRangeBreak,
		"../../testdata/loop/range_break.go",
		2,
	)
}
