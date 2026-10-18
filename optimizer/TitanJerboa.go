package optimizer

import (
	"Tokype/ast"
	"sync"
)

type Optimizer struct {
	env        interface{}
	cache      map[ast.Node]bool
	cacheMutex sync.RWMutex
	functions  map[string]*ast.FunctionStatement
	errors     []string
	varValues  map[string]*int64
}

func NewOptimizer() *Optimizer {
	return &Optimizer{
		cache:     make(map[ast.Node]bool),
		functions: make(map[string]*ast.FunctionStatement),
		errors:    []string{},
		varValues: make(map[string]*int64),
	}
}

func (o *Optimizer) Errors() []string {
	return o.errors
}

func (o *Optimizer) Optimize(program *ast.Program, baseDir string) *ast.Program {
	o.cache = make(map[ast.Node]bool)
	o.functions = make(map[string]*ast.FunctionStatement)
	o.varValues = make(map[string]*int64)

	visitedFiles = make(map[string]bool)
	program = o.resolveIncludes(program, baseDir)

	o.collectFunctions(program)
	o.collectVarValues(program)

	program = o.inlineFunctionCalls(program)
	program = o.optimizeCollections(program).(*ast.Program)

	o.constantFold(program)
	o.substituteForBounds(program)
	o.constantFold(program)

	program = o.foldDynamicLoops(program).(*ast.Program)

	program = o.foldFunctionCalls(program).(*ast.Program)

	o.optimizeControlFlow(program)
	o.removeDeadCode(program)

	program = o.removeEmptyFunctions(program)

	return program
}

func (o *Optimizer) collectFunctions(node ast.Node) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			o.collectFunctions(stmt)
		}
	case *ast.FunctionStatement:
		if n.Name != nil {
			o.functions[n.Name.Value] = n
		}
	}
}

func (o *Optimizer) collectVarValues(node ast.Node) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			o.collectVarValues(stmt)
		}
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			o.collectVarValues(stmt)
		}
	case *ast.FunctionStatement:
		if n.Body != nil {
			o.collectVarValues(n.Body)
		}
	case *ast.AssignStatement:
		if lit, ok := n.Value.(*ast.IntegerLiteral); ok {
			v := lit.Value
			o.varValues[n.Name.Value] = &v
		} else {
			o.varValues[n.Name.Value] = nil
		}
	}
}

func (o *Optimizer) mayBeZero(name string) bool {
	val, exists := o.varValues[name]
	if !exists {
		return true
	}
	if val == nil {
		return true
	}
	return *val == 0
}

func (o *Optimizer) Reset() {
	o.cacheMutex.Lock()
	defer o.cacheMutex.Unlock()
	o.cache = make(map[ast.Node]bool)
	o.functions = make(map[string]*ast.FunctionStatement)
	o.varValues = make(map[string]*int64)
}

func OptimizeAST(program *ast.Program, baseDir string) *ast.Program {
	optimizer := NewOptimizer()
	return optimizer.Optimize(program, baseDir)
}
