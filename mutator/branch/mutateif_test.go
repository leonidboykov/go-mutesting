package branch

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorIf(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorIf,
		"../../testdata/branch/mutateif.go",
		2,
	)
}
