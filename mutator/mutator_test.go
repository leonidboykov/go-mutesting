package mutator

import (
	"go/ast"
	"go/types"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockMutator(pkg *types.Package, info *types.Info, node ast.Node) []Mutation {
	// Do nothing

	return nil
}

func TestMockMutator(t *testing.T) {
	// Mock is not registered
	for _, name := range List() {
		if name == "mock" {
			assert.Fail(t, "mock should not be in the mutator list yet")
		}
	}

	m, err := New("mock")
	assert.Nil(t, m)
	assert.NotNil(t, err)

	// Register mock
	Register("mock", mockMutator)

	// Mock is registered
	found := slices.Contains(List(), "mock")
	assert.True(t, found)

	m, err = New("mock")
	assert.NotNil(t, m)
	assert.Nil(t, err)

	// Register mock a second time
	assert.Panics(t, func() { Register("mock", mockMutator) })

	// Register nil function
	assert.Panics(t, func() { Register("mockachino", nil) })
}
