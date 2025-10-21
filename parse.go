package mutesting

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	//nolint:staticcheck
	"golang.org/x/tools/go/packages"
)

// ParseFile parses the content of the given file and returns the corresponding [ast.File] node and its file set for positional information.
// If a fatal error is encountered the error return argument is not nil.
func ParseFile(file string) (*ast.File, *token.FileSet, error) {
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax}, file)
	if err != nil {
		return nil, nil, fmt.Errorf("load file: %w", err)
	}
	// TODO: this may lead to panic.
	return pkgs[0].Syntax[0], pkgs[0].Fset, err
}

func ParseAndTypeCheckFile(file string) (*ast.File, *token.FileSet, *types.Package, *types.Info, error) {
	fileAbs, _ := filepath.Abs(file)
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadSyntax,
	}, filepath.Dir(fileAbs))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("load file: %w", err)
	}
	packages.PrintErrors(pkgs)
	pkg := pkgs[0]
	var src *ast.File
	for _, f := range pkg.Syntax {
		if pkg.Fset.Position(f.Pos()).Filename == fileAbs {
			src = f
			break
		}
	}
	return src, pkg.Fset, pkg.Types, pkg.TypesInfo, nil
}

// Skips checks all comments and finds `//nomutesting` directives.
func Skips(fset *token.FileSet, f *ast.File) map[int]struct{} {
	skippedLines := make(map[int]struct{})
	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			text := strings.TrimLeft(comment.Text, "/ ")
			if strings.HasPrefix(text, "nomutesting") {
				skippedLines[fset.Position(comment.Pos()).Line] = struct{}{}
			}
		}
	}
	return skippedLines
}
