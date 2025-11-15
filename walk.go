package mutesting

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/leonidboykov/go-mutesting/mutator"
)

// PrintWalk traverses the AST of the given node and prints every node to STDOUT.
func PrintWalk(node ast.Node) {
	var level int
	ast.Inspect(node, func(node ast.Node) bool {
		if node != nil {
			level++
			fmt.Printf("%s(%p)%#v\n", strings.Repeat("\t", level), node, node)
		} else {
			level--
		}
		return true
	})
}

// MutateWalk mutates the given node with the given mutator returning a channel to control the mutation steps. It
// traverses the AST of the given node and calls the method Check of the given mutator to verify that a node can be
// mutated by the mutator. If a node can be mutated the method Mutate of the given mutator is executed with the node and
// the control channel. After completion of the traversal the control channel is closed.
func MutateWalk(pkg *packages.Package, node ast.Node, m mutator.Mutator, skippedLines map[int]struct{}) chan bool {
	w := &mutateWalk{
		changed: make(chan bool),
		mutator: m,
		pkg:     pkg,
		skipped: skippedLines,
	}

	go func() {
		ast.Walk(w, node)

		close(w.changed)
	}()

	return w.changed
}

type mutateWalk struct {
	changed chan bool
	mutator mutator.Mutator
	pkg     *packages.Package
	skipped map[int]struct{}
}

// Visit implements the Visit method of the ast.Visitor interface
func (w *mutateWalk) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return w
	}

	line := w.pkg.Fset.Position(node.Pos()).Line
	_, ok := w.skipped[line]
	if ok {
		return w
	}

	for _, m := range w.mutator(w.pkg.Types, w.pkg.TypesInfo, node) {
		m.Change()
		w.changed <- true
		<-w.changed

		m.Reset()
		w.changed <- true
		<-w.changed
	}

	return w
}
