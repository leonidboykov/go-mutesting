package arithmetic

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/leonidboykov/go-mutesting/mutator"
)

func init() {
	mutator.Register("arithmetic/assign_invert", MutatorArithmeticAssignInvert)
}

var assignInvertMutations = map[token.Token]token.Token{
	token.ADD_ASSIGN: token.SUB_ASSIGN,
	token.SUB_ASSIGN: token.ADD_ASSIGN,
	token.MUL_ASSIGN: token.QUO_ASSIGN,
	token.QUO_ASSIGN: token.MUL_ASSIGN,
	token.REM_ASSIGN: token.MUL_ASSIGN,
}

// MutatorArithmeticAssignInvert implements a mutator to invert change assign statements.
func MutatorArithmeticAssignInvert(_ *types.Package, info *types.Info, node ast.Node) []mutator.Mutation {
	n, ok := node.(*ast.AssignStmt)
	if !ok {
		return nil
	}

	// Skip concatenation case.
	if len(n.Lhs) > 0 && isConcatenationCase(info, n.Lhs[0]) {
		return nil
	}

	original := n.Tok
	mutated, ok := assignInvertMutations[n.Tok]
	if !ok {
		return nil
	}

	return []mutator.Mutation{
		{
			Change: func() {
				n.Tok = mutated
			},
			Reset: func() {
				n.Tok = original
			},
		},
	}
}
