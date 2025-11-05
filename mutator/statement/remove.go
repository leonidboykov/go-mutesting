package statement

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/leonidboykov/go-mutesting/internal/astutil"
	"github.com/leonidboykov/go-mutesting/mutator"
)

func init() {
	mutator.Register("statement/remove", MutatorRemoveStatement)
}

// MutatorRemoveStatement implements a mutator to remove statements.
func MutatorRemoveStatement(pkg *types.Package, info *types.Info, node ast.Node) []mutator.Mutation {
	var l []ast.Stmt

	switch n := node.(type) {
	case *ast.BlockStmt:
		l = n.List
	case *ast.CaseClause:
		l = n.Body
	}

	var mutations []mutator.Mutation

	for i, ni := range l {
		if checkRemoveStatement(ni) {
			li := i
			old := l[li]

			mutations = append(mutations, mutator.Mutation{
				Change: func() {
					l[li] = astutil.CreateNoopOfStatements(pkg, info, old)
				},
				Reset: func() {
					l[li] = old
				},
			})
		}
	}

	return mutations
}

func checkRemoveStatement(node ast.Stmt) bool {
	switch n := node.(type) {
	case *ast.AssignStmt:
		if n.Tok != token.DEFINE && !lhsIsBlank(n.Lhs) {
			return true
		}
	case *ast.ExprStmt, *ast.IncDecStmt:
		return true
	}

	return false
}

// lhsIsBlank checks if all expressions are blank. This allows to skip cases like in example:
//
//	_, _, _ = a, b, http.Header{}
func lhsIsBlank(lhs []ast.Expr) bool {
	for i := range lhs {
		if l, ok := lhs[i].(*ast.Ident); !ok || l.Name != "_" {
			return false
		}
	}
	return true
}
