package test

import (
	"bytes"
	"fmt"
	"go/printer"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/leonidboykov/go-mutesting"
	"github.com/leonidboykov/go-mutesting/mutator"
)

// Mutator tests a mutator.
// It mutates the given original file with the given mutator. Every mutation is then validated with the given changed file. The mutation overall count is validated with the given count.
func Mutator(t *testing.T, m mutator.Mutator, testFile string, count int) {
	// Test if mutator is not nil
	assert.NotNil(t, m)

	// Read the origianl source code
	data, err := os.ReadFile(testFile)
	assert.Nil(t, err)

	// Parse and type-check the original source code
	src, fset, pkg, info, err := mutesting.ParseAndTypeCheckFile(testFile)
	assert.Nil(t, err)

	skippedLines := mutesting.Skips(fset, src)

	// Mutate a non relevant node
	assert.Nil(t, m(pkg, info, src))

	// Count the actual mutations
	n := mutesting.CountWalk(pkg, info, fset, src, m, skippedLines)
	assert.Equal(t, count, n)

	// Mutate all relevant nodes -> test whole mutation process
	changed := mutesting.MutateWalk(pkg, info, fset, src, m, skippedLines)

	for i := 0; i < count; i++ {
		assert.True(t, <-changed)

		buf := new(bytes.Buffer)
		err = printer.Fprint(buf, fset, src)
		assert.Nil(t, err)

		changedFilename := fmt.Sprintf("%s.%d.go", testFile, i)
		changedFile, err := os.ReadFile(changedFilename)
		assert.Nil(t, err)

		if !assert.Equal(t, string(changedFile), buf.String(), fmt.Sprintf("For change file %q", changedFilename)) {
			err = os.WriteFile(fmt.Sprintf("%s.%d.go.new", testFile, i), buf.Bytes(), 0644)
			assert.Nil(t, err)
		}

		changed <- true

		assert.True(t, <-changed)

		buf = new(bytes.Buffer)
		err = printer.Fprint(buf, fset, src)
		assert.Nil(t, err)

		assert.Equal(t, string(data), buf.String())

		changed <- true
	}

	_, ok := <-changed
	assert.False(t, ok)
}
