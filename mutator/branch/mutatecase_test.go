package branch

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/test"
)

func TestMutatorCase(t *testing.T) {
	test.Mutator(
		t,
		MutatorCase,
		"../../testdata/branch/mutatecase.go",
		3,
	)
}
