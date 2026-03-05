package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// noOsExitAnalyzer reports direct calls to os.Exit in the main function
// of the main package. It encourages returning errors instead, which
// simplifies testing and composition of the application.
var noOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports direct calls to os.Exit in main.main",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// We only care about the main package.
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Only inspect the top-level main function.
			if fn.Recv != nil || fn.Name == nil || fn.Name.Name != "main" || fn.Body == nil {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				// Resolve the selector to a function symbol and ensure it is os.Exit.
				obj := pass.TypesInfo.Uses[sel.Sel]
				fnObj, ok := obj.(*types.Func)
				if !ok {
					return true
				}

				if fnObj.Pkg() != nil && fnObj.Pkg().Path() == "os" && fnObj.Name() == "Exit" {
					pass.Reportf(call.Lparen, "direct call to os.Exit in main.main is forbidden; return an error instead")
				}

				return true
			})

			// No need to re-visit inner declarations; we've already inspected the body.
			return false
		})
	}

	return nil, nil
}
