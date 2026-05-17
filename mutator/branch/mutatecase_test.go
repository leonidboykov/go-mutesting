package branch

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorCase(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorCase,
		"../../testdata/branch/mutatecase.go",
		3,
	)
}
