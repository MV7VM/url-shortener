// Package linter проверяет отсутствие log.fatal
package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// NoLogFatalAnalyzer reports calls to log.Fatal/Fatalf outside of main.main.
var NoLogFatalAnalyzer = &analysis.Analyzer{
	Name: "nologfatal",
	Doc:  "reports calls to log.Fatal/Fatalf outside main.main",
	Run:  runNoLogFatal,
}

func runNoLogFatal(pass *analysis.Pass) (interface{}, error) {
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
					sel, ok := x.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					obj := pass.TypesInfo.Uses[sel.Sel]
					fnObj, ok := obj.(*types.Func)
					if !ok {
						return true
					}
					if fnObj.Pkg() != nil && fnObj.Pkg().Path() == "log" &&
						(fnObj.Name() == "Fatal" || fnObj.Name() == "Fatalf") {
						if !insideMainMain() {
							pass.Reportf(x.Lparen, "log.%s is allowed only in main.main; return an error instead", fnObj.Name())
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
