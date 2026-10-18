package optimizer

import (
	"Tokype/ast"
	"Tokype/evaluator"
	"Tokype/value"
)

func (o *Optimizer) removeDeadCode(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			optimized := o.removeDeadCode(stmt)
			if !o.isDeadCode(optimized) {
				newStatements = append(newStatements, optimized)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			optimized := o.removeDeadCode(stmt)
			if !o.isDeadCode(optimized) {
				newStatements = append(newStatements, optimized)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.removeDeadCode(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.IfExpression:
		if o.hasUserFunctionCalls(n.Condition) {
			if n.Consequence != nil {
				n.Consequence = o.removeDeadCode(n.Consequence).(*ast.BlockStatement)
			}
			for i, elif := range n.Alternative {
				elif.Consequence = o.removeDeadCode(elif.Consequence).(*ast.BlockStatement)
				n.Alternative[i] = elif
			}
			if n.Else != nil {
				n.Else = o.removeDeadCode(n.Else).(*ast.BlockStatement)
			}
			return n
		}

		tempEnv := evaluator.NewEnvironment()
		result := evaluator.Eval(n.Condition, tempEnv)

		if result.Type == value.ValBool && !result.Bool && !o.hasVariables(n.Condition) {
			if n.Else != nil {
				return o.removeDeadCode(n.Else)
			}
			return &ast.BlockStatement{Statements: []ast.Node{}}
		}

		if n.Consequence != nil {
			n.Consequence = o.removeDeadCode(n.Consequence).(*ast.BlockStatement)
		}
		for i, elif := range n.Alternative {
			elif.Consequence = o.removeDeadCode(elif.Consequence).(*ast.BlockStatement)
			n.Alternative[i] = elif
		}
		if n.Else != nil {
			n.Else = o.removeDeadCode(n.Else).(*ast.BlockStatement)
		}
		return n

	case *ast.ForStatement:
		if condition, ok := n.Condition.(*ast.InfixExpression); ok {
			if !o.hasUserFunctionCalls(condition.Right) {
				tempEnv := evaluator.NewEnvironment()
				initResult := evaluator.Eval(n.Initialization.Value, tempEnv)
				endResult := evaluator.Eval(condition.Right, tempEnv)

				if !initResult.IsNil() && !endResult.IsNil() && !o.hasVariables(n.Initialization.Value) && !o.hasVariables(condition.Right) {
					initVal, _ := initResult.ToFloat64()
					endVal, _ := endResult.ToFloat64()

					if initVal > endVal {
						return &ast.BlockStatement{Statements: []ast.Node{}}
					}
				}
			}
		}

		if n.Body != nil {
			n.Body = o.removeDeadCode(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.WhileStatement:
		if o.hasUserFunctionCalls(n.Condition) {
			if n.Body != nil {
				n.Body = o.removeDeadCode(n.Body).(*ast.BlockStatement)
			}
			return n
		}

		tempEnv := evaluator.NewEnvironment()
		result := evaluator.Eval(n.Condition, tempEnv)

		if result.Type == value.ValBool && !result.Bool && !o.hasVariables(n.Condition) {
			return &ast.BlockStatement{Statements: []ast.Node{}}
		}

		if n.Body != nil {
			n.Body = o.removeDeadCode(n.Body).(*ast.BlockStatement)
		}
		return n

	default:
		return n
	}
}

func (o *Optimizer) isDeadCode(node ast.Node) bool {
	if node == nil {
		return true
	}

	if call, ok := node.(*ast.CallExpression); ok {
		if ident, ok := call.Function.(*ast.Identifier); ok {
			if o.hasSideEffects(ident.Value) {
				return false
			}
		}
	}

	switch n := node.(type) {
	case *ast.BlockStatement:
		return len(n.Statements) == 0
	case *ast.Program:
		return len(n.Statements) == 0
	default:
		return false
	}
}
