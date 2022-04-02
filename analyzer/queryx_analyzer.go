package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var QueryxAnalyzer = &analysis.Analyzer{
	Name:     "rdbxqueryx",
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

		for _, b := range body {
			ifStmt, ok := b.(*ast.IfStmt)
			if !ok {
				continue
			}

			aStmt, ok := ifStmt.Init.(*ast.AssignStmt)
			if !ok {
				continue
			}

			xCall, ok := aStmt.Rhs[0].(*ast.CallExpr)
			if !ok {
				continue
			}

			fun, ok := xCall.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			if methName := fun.Sel.Name; methName != "Queryx" {
				continue
			}

			_, ok = xCall.Args[2].(*ast.UnaryExpr)
			if !ok {
				pass.Reportf(node.Pos(), "non-pointer model is passed")
				continue
			}
		}
	})

	return nil, nil
}
