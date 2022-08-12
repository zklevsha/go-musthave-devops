// Package checker implements custom static checkers
package checker

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func checkExitCalls(f *ast.FuncDecl, pass *analysis.Pass) {
	ast.Inspect(f, func(n ast.Node) bool {
		fc, ok := n.(*ast.CallExpr)
		if ok {
			if fun, ok := fc.Fun.(*ast.SelectorExpr); ok {
				funcName := fun.Sel.Name
				if strings.Contains("os.Exit", funcName) {
					pass.Reportf(fc.Pos(), "found os.Exit call in main()")
				}

			}
		}
		return true
	})
}

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "check for os.Exit() calls inside main() function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			// searching for main function
			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				if funcDecl.Name.Name == "main" {
					checkExitCalls(funcDecl, pass)
				}
			}
			return true
		})
	}
	return nil, nil
}
