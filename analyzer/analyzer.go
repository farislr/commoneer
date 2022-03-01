package analyzer

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "Queryx",
	Doc:      "check queryx model to be pointer",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		body := funcDecl.Body.List

		// fmt.Printf("funcDecl: %v\n", funcDecl)
		for _, b := range body {
			ifStmt, ok := b.(*ast.IfStmt)
			if !ok {
				continue
			}

			fmt.Printf("ifStmt: %v\n", ifStmt)

		}

	})

	return nil, nil
}
