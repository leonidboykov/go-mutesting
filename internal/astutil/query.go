package astutil

import (
	"go/ast"
	"go/token"
	"go/types"
)

// IdentifiersInStatement returns all identifiers with their found in a statement.
func IdentifiersInStatement(pkg *types.Package, info *types.Info, stmt ast.Stmt) []ast.Expr {
	var identifiers []ast.Expr

	ast.PreorderStack(stmt, nil, func(node ast.Node, _ []ast.Node) bool {
		switch n := node.(type) {
		case *ast.Ident:
			// Ignore the blank identifier
			if n.Name == "_" {
				return false
			}

			// Ignore keywords
			if token.Lookup(n.Name) != token.IDENT {
				return false
			}

			// We are only interested in variables
			if obj, ok := info.Uses[n]; ok {
				if _, ok := obj.(*types.Var); !ok {
					return false
				}
			}

			// FIXME instead of manually creating a new node, clone it and trim the node from its comments and position https://github.com/zimmski/go-mutesting/issues/49
			identifiers = append(identifiers, &ast.Ident{
				Name: n.Name,
			})

			return false
		case *ast.SelectorExpr:
			if !checkForSelectorExpr(n) {
				return false
			}

			// Check if we need to instantiate the expression
			initialize := false
			if n.Sel != nil {
				if obj, ok := info.Uses[n.Sel]; ok {
					t := obj.Type()

					switch t.Underlying().(type) {
					case *types.Array, *types.Map, *types.Slice, *types.Struct:
						initialize = true
					}
				}
			}

			if initialize {
				// FIXME we need to clone the node and trim comments and position recursively https://github.com/zimmski/go-mutesting/issues/49
				identifiers = append(identifiers, &ast.CompositeLit{
					Type: n,
				})
			} else {
				// FIXME we need to clone the node and trim comments and position recursively https://github.com/zimmski/go-mutesting/issues/49
				identifiers = append(identifiers, n)
			}

			return false
		}

		return true
	})

	return identifiers
}

func checkForSelectorExpr(node ast.Expr) bool {
	switch n := node.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return checkForSelectorExpr(n.X)
	}

	return false
}

// Functions returns all found functions.
func Functions(n ast.Node) []*ast.FuncDecl {
	var functions []*ast.FuncDecl

	ast.PreorderStack(n, nil, func(node ast.Node, stack []ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			functions = append(functions, n)

			return false
		}
		return true
	})

	return functions
}
