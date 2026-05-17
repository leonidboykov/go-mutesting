package arithmetic

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorArithmeticBase(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorArithmeticBase,
		"../../testdata/arithmetic/base.go",
		5,
	)
}
