package optimizer

import (
	"Tokype/ast"
	"Tokype/lexer"
	"Tokype/parser"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var visitedFiles = make(map[string]bool)

func (o *Optimizer) resolveIncludes(program *ast.Program, baseDir string) *ast.Program {
	newStatements := []ast.Node{}

	for _, stmt := range program.Statements {
		if inc, ok := stmt.(*ast.IncludeStatement); ok {
			if !strings.HasSuffix(inc.Path, ".tp") {
				newStatements = append(newStatements, inc)
				continue
			}

			var fullPath string
			if filepath.IsAbs(inc.Path) {
				fullPath = inc.Path
			} else {
				fullPath = filepath.Join(baseDir, inc.Path)
			}

			if visitedFiles[fullPath] {
				continue
			}
			visitedFiles[fullPath] = true

			content, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Printf("Error: cannot include %s: %v\n", inc.Path, err)
				continue
			}

			l := lexer.NewLexer(string(content))
			p := parser.NewParser(l)
			includedProg := p.ParseProgram()

			if len(p.Errors()) > 0 {
				fmt.Printf("Parse errors in %s:\n", inc.Path)
				for _, e := range p.Errors() {
					fmt.Println("  " + e)
				}
				continue
			}

			includedProg = o.resolveIncludes(includedProg, filepath.Dir(fullPath))

			if inc.Namespace != "" {
				prefixFunctions(includedProg, inc.Namespace)
			}

			newStatements = append(newStatements, includedProg.Statements...)
		} else {
			newStatements = append(newStatements, stmt)
		}
	}

	program.Statements = newStatements
	return program
}

func prefixFunctions(program *ast.Program, namespace string) {
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			if s.Name != nil {
				s.Name.Value = namespace + "_" + s.Name.Value
			}
		case *ast.AssignStatement:
			if s.Name != nil {
				s.Name.Value = namespace + "_" + s.Name.Value
			}
		}
	}
}
