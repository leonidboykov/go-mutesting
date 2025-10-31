package mutesting

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	//nolint:staticcheck
	"golang.org/x/tools/go/packages"
)

// ParseFile parses the content of the given file and returns the corresponding [ast.File] node and its file set for
// positional information. If a fatal error is encountered the error return argument is not nil.
func ParseFile(filename string) (*ast.File, *token.FileSet, error) {
	pkg, src, err := parseFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("parse file: %w", err)
	}
	return src, pkg.Fset, err
}

func ParseAndTypeCheckFile(filename string) (*ast.File, *packages.Package, error) {
	pkg, src, err := parseFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("parse file: %w", err)
	}
	return src, pkg, nil
}

// Skips checks all comments and finds `//nomutesting` directives.
func Skips(fset *token.FileSet, src *ast.File) map[int]struct{} {
	skippedLines := make(map[int]struct{})
	for _, commentGroup := range src.Comments {
		for _, comment := range commentGroup.List {
			text := strings.TrimLeft(comment.Text, "/ ")
			if strings.HasPrefix(text, "nomutesting") {
				skippedLines[fset.Position(comment.Pos()).Line] = struct{}{}
			}
		}
	}
	return skippedLines
}

func parseFile(filename string) (*packages.Package, *ast.File, error) {
	filenameAbs, err := filepath.Abs(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("get abs filename: %w", err)
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadSyntax,
	}, filepath.Dir(filenameAbs))
	if err != nil {
		return nil, nil, fmt.Errorf("load package: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, nil, errors.New("no packages")
	}

	pkg := pkgs[0]

	var src *ast.File
	for _, f := range pkg.Syntax {
		if pkg.Fset.Position(f.Pos()).Filename == filenameAbs {
			src = f
			break
		}
	}
	if src == nil {
		return nil, nil, errors.New("syntax file not found")
	}
	return pkg, src, nil
}
