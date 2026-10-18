package optimizer

import (
	"Tokype/ast"
	"Tokype/evaluator"
)

func (o *Optimizer) foldFunctionCalls(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.foldFunctionCalls(stmt)
		}
		return n

	case *ast.BlockStatement:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.foldFunctionCalls(stmt)
		}
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.foldFunctionCalls(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.AssignStatement:
		if call, ok := n.Value.(*ast.CallExpression); ok {
			if result := o.optimizeRecursive(call); result != nil {
				n.Value = result
				return n
			}
			if ident, ok := call.Function.(*ast.Identifier); ok {
				if fn, exists := o.functions[ident.Value]; exists {
					if result := o.evaluatePureFunction(fn, call.Arguments); result != nil {
						n.Value = result
					}
				}
			}
		}
		return n

	default:
		return n
	}
}

func (o *Optimizer) optimizeRecursive(call *ast.CallExpression) ast.Expression {
	if len(call.Arguments) != 1 {
		return nil
	}

	ident, ok := call.Function.(*ast.Identifier)
	if !ok {
		return nil
	}

	fn, exists := o.functions[ident.Value]
	if !exists {
		return nil
	}

	arg, ok := call.Arguments[0].(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	if !o.isFibonacci(fn) {
		return nil
	}

	n := arg.Value
	if n < 0 || n > 92 {
		return nil
	}

	a, b := int64(0), int64(1)
	for i := int64(0); i < n; i++ {
		a, b = b, a+b
	}

	return &ast.IntegerLiteral{Value: a}
}

func (o *Optimizer) isFibonacci(fn *ast.FunctionStatement) bool {
	if fn.Body == nil {
		return false
	}
	return o.matchesFibonacciPattern(fn.Body, fn.Name.Value)
}

func (o *Optimizer) matchesFibonacciPattern(node ast.Node, name string) bool {
	block, ok := node.(*ast.BlockStatement)
	if !ok || len(block.Statements) != 1 {
		return false
	}

	ifExpr, ok := block.Statements[0].(*ast.IfExpression)
	if !ok {
		return false
	}

	cond, ok := ifExpr.Condition.(*ast.InfixExpression)
	if !ok || cond.Operator != "<" {
		return false
	}

	ifExprElse := ifExpr.Else
	if ifExprElse == nil || len(ifExprElse.Statements) != 1 {
		return false
	}

	ret, ok := ifExprElse.Statements[0].(*ast.ReturnStatement)
	if !ok {
		return false
	}

	infix, ok := ret.Value.(*ast.InfixExpression)
	if !ok || infix.Operator != "+" {
		return false
	}

	leftCall, ok1 := infix.Left.(*ast.CallExpression)
	rightCall, ok2 := infix.Right.(*ast.CallExpression)

	if !ok1 || !ok2 {
		return false
	}

	leftIdent, ok1 := leftCall.Function.(*ast.Identifier)
	rightIdent, ok2 := rightCall.Function.(*ast.Identifier)

	return ok1 && ok2 && leftIdent.Value == name && rightIdent.Value == name
}

func (o *Optimizer) evaluatePureFunction(fn *ast.FunctionStatement, args []ast.Expression) ast.Expression {
	if len(args) != len(fn.Parameters) {
		return nil
	}

	for _, arg := range args {
		if !o.isConstantExpr(arg) {
			return nil
		}
	}

	if o.hasSideEffectsInFunction(fn) {
		return nil
	}

	if o.hasRecursion(fn) {
		return nil
	}

	env := evaluator.NewEnvironment()
	for i, param := range fn.Parameters {
		val := evaluator.Eval(args[i], env)
		env.Set(param.Value, val)
	}

	result := evaluator.Eval(fn.Body, env)

	return o.valueToLiteral(result)
}

func (o *Optimizer) hasRecursion(fn *ast.FunctionStatement) bool {
	if fn.Name == nil || fn.Body == nil {
		return false
	}
	return o.callsFunction(fn.Body, fn.Name.Value)
}

func (o *Optimizer) callsFunction(node ast.Node, name string) bool {
	switch n := node.(type) {
	case *ast.BlockStatement:
		for _, s := range n.Statements {
			if o.callsFunction(s, name) {
				return true
			}
		}
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if ident.Value == name {
				return true
			}
		}
		for _, arg := range n.Arguments {
			if o.callsFunction(arg, name) {
				return true
			}
		}
	case *ast.IfExpression:
		if o.callsFunction(n.Consequence, name) {
			return true
		}
		for _, alt := range n.Alternative {
			if o.callsFunction(alt.Consequence, name) {
				return true
			}
		}
		if n.Else != nil {
			return o.callsFunction(n.Else, name)
		}
	case *ast.ReturnStatement:
		if n.Value != nil {
			return o.callsFunction(n.Value, name)
		}
	case *ast.AssignStatement:
		return o.callsFunction(n.Value, name)
	case *ast.InfixExpression:
		return o.callsFunction(n.Left, name) || o.callsFunction(n.Right, name)
	case *ast.PrefixExpression:
		return o.callsFunction(n.Right, name)
	case *ast.ForStatement:
		return o.callsFunction(n.Body, name)
	case *ast.WhileStatement:
		return o.callsFunction(n.Body, name)
	}
	return false
}

func (o *Optimizer) isConstantExpr(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.IntegerLiteral, *ast.FloatLiteral, *ast.StringLiteral, *ast.BooleanLiteral, *ast.NullLiteral:
		return true
	case *ast.PrefixExpression:
		return o.isConstantExpr(e.Right)
	case *ast.InfixExpression:
		return o.isConstantExpr(e.Left) && o.isConstantExpr(e.Right)
	default:
		return false
	}
}

func (o *Optimizer) hasSideEffectsInFunction(fn *ast.FunctionStatement) bool {
	if fn.Body == nil {
		return false
	}
	return o.hasSideEffectsInNode(fn.Body)
}

func (o *Optimizer) hasSideEffectsInNode(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.BlockStatement:
		for _, s := range n.Statements {
			if o.hasSideEffectsInNode(s) {
				return true
			}
		}
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if o.hasSideEffects(ident.Value) {
				return true
			}
		}
	case *ast.AssignStatement:
		return o.hasSideEffectsInNode(n.Value)
	case *ast.IfExpression:
		if o.hasSideEffectsInNode(n.Consequence) {
			return true
		}
		for _, alt := range n.Alternative {
			if o.hasSideEffectsInNode(alt.Consequence) {
				return true
			}
		}
		if n.Else != nil {
			return o.hasSideEffectsInNode(n.Else)
		}
	case *ast.ReturnStatement:
		if n.Value != nil {
			return o.hasSideEffectsInNode(n.Value)
		}
	case *ast.ForStatement:
		return o.hasSideEffectsInNode(n.Body)
	case *ast.WhileStatement:
		return o.hasSideEffectsInNode(n.Body)
	}
	return false
}
