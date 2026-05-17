package numbers

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorNumbersIncrementer(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorNumbersIncrementer,
		"../../testdata/numbers/incrementer.go",
		2,
	)
}
