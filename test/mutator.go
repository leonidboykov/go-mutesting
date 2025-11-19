package test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"

	"github.com/leonidboykov/go-mutesting"
	"github.com/leonidboykov/go-mutesting/internal/importing"
	"github.com/leonidboykov/go-mutesting/mutator"
)

// Mutator tests a mutator.
// It mutates the given original file with the given mutator. Every mutation is then validated with the given changed
// file. The mutation overall count is validated with the given count.
func Mutator(t *testing.T, m mutator.Mutator, testFile string, expectedMutationCount int) {
	t.Helper()

	// Test if mutator is not nil
	assert.NotNil(t, m)

	// Read the origianl source code
	data, err := os.ReadFile(testFile)
	assert.Nil(t, err)

	// Parse and type-check the original source code
	src, pkg, err := importing.ParseAndTypeCheckFile(testFile)
	assert.Nil(t, err)

	skippedLines := importing.Skips(pkg.Fset, src)

	// Mutate a non relevant node
	assert.Nil(t, m(pkg.Types, pkg.TypesInfo, src))

	// Count the actual mutations
	n := countWalk(pkg, src, m, skippedLines)
	assert.Equal(t, expectedMutationCount, n)

	var mutationsCount int
	// Mutate all relevant nodes -> test whole mutation process
	mutesting.MutateWalk(pkg, src, m, skippedLines,
		func() {
			buf := new(bytes.Buffer)
			err = printer.Fprint(buf, pkg.Fset, src)
			assert.Nil(t, err)

			dir := filepath.Dir(testFile)
			fname := filepath.Base(testFile)
			changedFilename := fmt.Sprintf("%s/_%s.%d.go", dir, fname, mutationsCount)
			changedFile, err := os.ReadFile(changedFilename)
			assert.Nil(t, err)

			if !assert.Equal(t, string(changedFile), buf.String(), fmt.Sprintf("For change file %q", changedFilename)) {
				err = os.WriteFile(fmt.Sprintf("%s.%d.go.new", testFile, mutationsCount), buf.Bytes(), 0644)
				assert.Nil(t, err)
			}

			mutationsCount++
		},
		func() {
			buf := new(bytes.Buffer)
			err = printer.Fprint(buf, pkg.Fset, src)
			assert.Nil(t, err)

			assert.Equal(t, string(data), buf.String())
		},
	)

	assert.Equal(t, expectedMutationCount, mutationsCount)
}

// countWalk returns the number of corresponding mutations for a given mutator. It traverses the AST of the given node
// and calls the method Check of the given mutator for every node and sums up the returned counts. After completion of
// the traversal the final counter is returned.
func countWalk(pkg *packages.Package, node ast.Node, m mutator.Mutator, skippedLines map[int]struct{}) int {
	var count int
	for n := range ast.Preorder(node) {
		line := pkg.Fset.Position(node.Pos()).Line
		if _, ok := skippedLines[line]; !ok {
			count += len(m(pkg.Types, pkg.TypesInfo, n))
		}
	}
	return count
}
