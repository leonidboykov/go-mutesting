package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareStrings(t *testing.T) {
	t.Parallel()

	const expectedDiff = `--- Original
+++ Mutation: mutator name
@@ -1 +1 @@
-Hello
+World` + "\n\n"

	diff, err := CompareStrings("Hello", "World", "mutator name")
	assert.NoError(t, err)
	assert.Equal(t, expectedDiff, diff)
}
