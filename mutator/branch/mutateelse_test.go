package branch

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorElse(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorElse,
		"../../testdata/branch/mutateelse.go",
		1,
	)
}
