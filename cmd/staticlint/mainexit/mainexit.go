// Package mainexit contains os.exit call checker
package mainexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name:     "mainexit",
	Doc:      "Checker reports os.Exit calls inside main func of a main package.",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(
			file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.FuncDecl:
					return x.Name.Name == "main"
				case *ast.SelectorExpr:
					ident, ok := x.X.(*ast.Ident)
					if !ok || ident.Name != "os" || x.Sel.Name != "Exit" {
						return false
					}

					pass.Reportf(x.Pos(), "main must not contain os.Exit call")
				}
				return true
			},
		)
	}

	return nil, nil
}
