// Package linter проверяет отсутствие os.exit
package linter

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer reports direct calls to os.Exit in the main function
// of the main package. It encourages returning errors instead, which
// simplifies testing and composition of the application.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports direct calls to os.Exit in main.main",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// We only care about the main package.
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Получаем путь к текущему пакету (ваш код)
	currentPkgPath := pass.Pkg.Path()

	// Если это тестовый пакет или стандартная библиотека - пропускаем
	if strings.Contains(currentPkgPath, ".test") ||
		strings.HasPrefix(currentPkgPath, "go/") ||
		strings.HasPrefix(currentPkgPath, "golang.org/") {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Проверяем, что файл принадлежит нашему пакету
		filename := pass.Fset.File(file.Pos()).Name()

		// Пропускаем сгенерированные файлы
		if strings.Contains(filename, ".gen.") ||
			strings.HasSuffix(filename, "_test.go") {
			continue
		}

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

				// Проверяем, что вызов os.Exit происходит из нашего кода,
				// а не из импортированной библиотеки
				if fnObj.Pkg() != nil && fnObj.Pkg().Path() == "os" && fnObj.Name() == "Exit" {
					// Дополнительно проверяем позицию вызова
					callPos := pass.Fset.Position(call.Pos())

					// Если файл не из нашего проекта (например, из vendor) - пропускаем
					if strings.Contains(callPos.Filename, "/vendor/") ||
						strings.Contains(callPos.Filename, "/pkg/mod/") {
						return true
					}

					pass.Reportf(call.Lparen, "direct call to os.Exit in main.main is forbidden; return an error instead")
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
