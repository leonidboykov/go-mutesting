package expression

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorRemoveTerm(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorRemoveTerm,
		"../../testdata/expression/remove.go",
		6,
	)
}
