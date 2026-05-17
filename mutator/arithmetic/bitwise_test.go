package arithmetic

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorArithmeticBitwise(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorArithmeticBitwise,
		"../../testdata/arithmetic/bitwise.go",
		6,
	)
}
