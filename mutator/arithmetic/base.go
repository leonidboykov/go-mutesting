package arithmetic

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/leonidboykov/go-mutesting/mutator"
)

func init() {
	mutator.Register("arithmetic/base", MutatorArithmeticBase)
}

var arithmeticMutations = map[token.Token]token.Token{
	token.ADD: token.SUB,
	token.SUB: token.ADD,
	token.MUL: token.QUO,
	token.QUO: token.MUL,
	token.REM: token.MUL,
}

// MutatorArithmeticBase implements a mutator to change base arithmetic.
func MutatorArithmeticBase(_ *types.Package, info *types.Info, node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	// Skip concatenation case.
	if isConcatenationCase(info, n) {
		return nil
	}

	original := n.Op
	mutated, ok := arithmeticMutations[n.Op]
	if !ok {
		return nil
	}

	return []mutator.Mutation{
		{
			Change: func() {
				n.Op = mutated
			},
			Reset: func() {
				n.Op = original
			},
		},
	}
}

// isConcatenationCase checks if operation is a concatenation case. Mutation leads syntax error.
func isConcatenationCase(info *types.Info, expr ast.Expr) bool {
	if x, ok := info.Types[expr]; ok {
		if t, ok := x.Type.Underlying().(*types.Basic); ok {
			return t.Kind() == types.String || t.Kind() == types.UntypedString
		}
	}
	return false
}
