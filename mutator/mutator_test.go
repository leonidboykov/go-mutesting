package mutator

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockMutator(pkg *types.Package, info *types.Info, node ast.Node) []Mutation {
	return nil // Do nothing
}

func TestMockMutator(t *testing.T) {
	// Mock is not registered
	require.NotContains(t, List(), "mock", "mock should not be in the mutator list yet")

	m, err := New("mock")
	assert.Nil(t, m)
	assert.NotNil(t, err)

	// Register mock
	assert.NotPanics(t, func() { Register("mock", mockMutator) })

	// Mock is registered
	assert.Contains(t, List(), "mock")

	m, err = New("mock")
	assert.NotNil(t, m)
	require.NoError(t, err)

	// Register mock a second time
	assert.PanicsWithValue(t, `mutator "mock" already registered`, func() { Register("mock", mockMutator) })

	// Register nil function
	assert.PanicsWithValue(t, "mutator function is nil", func() { Register("mockachino", nil) })
}
