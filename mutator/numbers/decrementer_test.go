package numbers

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorNumbersDecrementer(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorNumbersDecrementer,
		"../../testdata/numbers/decrementer.go",
		2,
	)
}
