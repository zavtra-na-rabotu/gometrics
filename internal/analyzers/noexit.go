// Package analyzers is a package for code analyzers
package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name: "noexitcalls",
	Doc:  "check for os.Exit calls in main function of main package",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Check if package is "main"
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	// Loop over all files in package
	for _, file := range pass.Files {
		// Look for main method
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name.Name != "main" {
				continue
			}

			// Check main method for os.Exit calls
			if funcDecl.Body != nil {
				checkForOsExit(pass, funcDecl.Body)
			}
		}
	}
	return nil, nil
}

// checkForOsExit is recursive function to search for os.Exit calls
func checkForOsExit(pass *analysis.Pass, block *ast.BlockStmt) {
	for _, stmt := range block.List {
		switch s := stmt.(type) {
		case *ast.ExprStmt:
			if call, ok := s.X.(*ast.CallExpr); ok {
				if isOsExitCall(call) {
					pass.Reportf(call.Pos(), "using os.Exit in the main function of the main package is prohibited")
				}
			}
		case *ast.IfStmt:
			// Check if-else blocks
			checkForOsExit(pass, s.Body)
			if s.Else != nil {
				if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
					checkForOsExit(pass, elseBlock)
				}
			}
		case *ast.ForStmt:
			// Check for loops
			checkForOsExit(pass, s.Body)
		case *ast.SwitchStmt:
			// Check switch statements
			for _, stmt := range s.Body.List {
				if caseClause, ok := stmt.(*ast.CaseClause); ok {
					for _, bodyStmt := range caseClause.Body {
						if blockStmt, ok := bodyStmt.(*ast.BlockStmt); ok {
							checkForOsExit(pass, blockStmt)
						}
					}
				}
			}
		}
	}
}

// isOsExitCall is to check if expression is os.Exit
func isOsExitCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
			return true
		}
	}
	return false
}
