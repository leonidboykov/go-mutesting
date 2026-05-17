package conditional

import (
	"testing"

	"github.com/leonidboykov/go-mutesting/internal/mutatortest"
)

func TestMutatorConditionalNegated(t *testing.T) {
	mutatortest.Run(
		t,
		MutatorConditionalNegated,
		"../../testdata/conditional/negated.go",
		6,
	)
}
