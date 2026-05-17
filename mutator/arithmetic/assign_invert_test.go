package arithmetic

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorArithmeticAssignInvert(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorArithmeticAssignInvert,
		"../../testdata/arithmetic/assign_invert.go",
		5,
	)
}
