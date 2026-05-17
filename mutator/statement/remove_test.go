package statement

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorRemoveStatement(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorRemoveStatement,
		"../../testdata/statement/remove.go",
		17,
	)
}
