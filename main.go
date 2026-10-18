package main

import (
	"Tokype/evaluator"
	"Tokype/lexer"
	"Tokype/optimizer"
	"Tokype/parser"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
)

const Version = "0.0.3 (unstable)"

func main() {
	debug.SetMemoryLimit(50 * 1024 * 1024)
	debug.SetGCPercent(20)
	if len(os.Args) < 2 {
		fmt.Println("Usage: tokype <file.tp>")
		return
	}

	if os.Args[1] == "version" {
		fmt.Printf("Tokype version %s\n", Version)
		return
	}
	if os.Args[1] == "help" {
		fmt.Println("tokype <file> - run tokype script")
		fmt.Println("tokype help         - show this help")
		fmt.Println("tokype version      - show version")
		fmt.Println("tokype license      - show license")
		fmt.Println("tokype ast <file>   - show AST without execution")
		return
	}
	if os.Args[1] == "license" {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Println("Error retrieving program path:", err)
			return
		}

		exeDir := filepath.Dir(exePath)
		licensePath := filepath.Join(exeDir, "License")

		data, err := os.ReadFile(licensePath)
		if err != nil {
			fmt.Println("Error reading license file:", err)
			return
		}
		fmt.Println(string(data))
		return
	}
	if os.Args[1] == "ast" {
		if len(os.Args) < 3 {
			fmt.Println("Usage: tokype ast <file.tp>")
			return
		}

		filePath := os.Args[2]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		code := string(data)

		l := lexer.NewLexer(code)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Println("Parser errors:")
			for _, err := range p.Errors() {
				fmt.Printf("  %s\n", err)
			}
			return
		}
		optimizedProgram := optimizer.OptimizeAST(program, filepath.Dir(os.Args[1]))

		fmt.Println(optimizedProgram.String())
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	code := string(data)

	l := lexer.NewLexer(code)
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		return
	}

	env := evaluator.NewEnvironment()
	opt := optimizer.NewOptimizer()
	optimizedProgram := optimizer.OptimizeAST(program, filepath.Dir(os.Args[1]))

	if len(opt.Errors()) > 0 {
		fmt.Println("Optimizer errors:")
		for _, e := range opt.Errors() {
			fmt.Println("  " + e)
		}
		return
	}

	result := evaluator.Eval(optimizedProgram, env)

	if !result.IsNil() {
		fmt.Printf("Result: %v\n", result.String())
	}
}
