// Package linter проверяет отсутствие panic
package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// NoPanicAnalyzer reports direct calls to panic outside of main.main.
var NoPanicAnalyzer = &analysis.Analyzer{
	Name: "nopanic",
	Doc:  "reports direct calls to panic outside main.main",
	Run:  runNoPanic,
}

func runNoPanic(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		var stack []bool
		insideMainMain := func() bool {
			if len(stack) == 0 {
				return false
			}
			return stack[len(stack)-1]
		}
		var inspect func(ast.Node)
		inspect = func(node ast.Node) {
			ast.Inspect(node, func(n ast.Node) bool {
				if n == nil {
					return true
				}
				switch x := n.(type) {
				case *ast.FuncDecl:
					isMain := pass.Pkg != nil &&
						pass.Pkg.Name() == "main" &&
						x.Recv == nil &&
						x.Name != nil &&
						x.Name.Name == "main"
					stack = append(stack, isMain)
					if x.Body != nil {
						inspect(x.Body)
					}
					stack = stack[:len(stack)-1]
					return false
				case *ast.FuncLit:
					stack = append(stack, false)
					if x.Body != nil {
						inspect(x.Body)
					}
					stack = stack[:len(stack)-1]
					return false
				case *ast.CallExpr:
					ident, ok := x.Fun.(*ast.Ident)
					if !ok {
						return true
					}
					if ident.Name == "panic" {
						if !insideMainMain() {
							pass.Reportf(x.Lparen, "panic is allowed only in main.main; return an error instead")
						}
					}
					return true
				}
				return true
			})
		}
		inspect(file)
	}
	return nil, nil
}
