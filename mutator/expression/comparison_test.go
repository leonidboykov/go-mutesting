package expression

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorComparison(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorComparison,
		"../../testdata/expression/comparison.go",
		4,
	)
}
