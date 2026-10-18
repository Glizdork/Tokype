package optimizer

import (
	"Tokype/ast"
	"Tokype/evaluator"
)

func (o *Optimizer) constantFold(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.constantFold(stmt)
		}
		return n

	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			folded := o.constantFold(stmt)
			if folded != nil {
				newStatements = append(newStatements, folded)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.constantFold(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.AssignStatement:
		if n.Value != nil {
			n.Value = o.constantFoldExpression(n.Value)

			if ident, ok := n.Value.(*ast.Identifier); ok && ident.Value == n.Name.Value {
				return nil
			}

			if infix, ok := n.Value.(*ast.InfixExpression); ok {
				if leftId, ok := infix.Left.(*ast.Identifier); ok && leftId.Value == n.Name.Value {
					if rightInt, ok := infix.Right.(*ast.IntegerLiteral); ok {
						if (infix.Operator == "+" || infix.Operator == "-") && rightInt.Value == 0 {
							return nil
						}
						if (infix.Operator == "*" || infix.Operator == "/") && rightInt.Value == 1 {
							return nil
						}
						if infix.Operator == "*" && rightInt.Value == 0 {
							n.Value = &ast.IntegerLiteral{Value: 0}
							return n
						}
					}
					if infix.Operator == "-" {
						if rightId, ok := infix.Right.(*ast.Identifier); ok && rightId.Value == n.Name.Value {
							n.Value = &ast.IntegerLiteral{Value: 0}
							return n
						}
					}
				}
			}

			if folded := o.evaluateConstant(n.Value); folded != nil {
				n.Value = folded
			}
		}
		return n

	case *ast.InfixExpression:
		n.Left = o.constantFoldExpression(n.Left)
		n.Right = o.constantFoldExpression(n.Right)

		if n.Operator == "-" {
			if leftId, ok1 := n.Left.(*ast.Identifier); ok1 {
				if rightId, ok2 := n.Right.(*ast.Identifier); ok2 {
					if leftId.Value == rightId.Value {
						return &ast.IntegerLiteral{Value: 0}
					}
				}
			}
		}

		if n.Operator == "/" {
			if leftId, ok1 := n.Left.(*ast.Identifier); ok1 {
				if rightId, ok2 := n.Right.(*ast.Identifier); ok2 {
					if leftId.Value == rightId.Value && !o.mayBeZero(leftId.Value) {
						return &ast.IntegerLiteral{Value: 1}
					}
				}
			}

			if leftInfix, ok := n.Left.(*ast.InfixExpression); ok && leftInfix.Operator == "*" {
				if leftId, ok := leftInfix.Left.(*ast.Identifier); ok {
					if midId, ok := leftInfix.Right.(*ast.Identifier); ok {
						if rightId, ok := n.Right.(*ast.Identifier); ok {
							if midId.Value == rightId.Value && !o.mayBeZero(midId.Value) {
								return leftId
							}
						}
					}
				}
				if leftId, ok := leftInfix.Left.(*ast.Identifier); ok {
					if midId, ok := leftInfix.Right.(*ast.Identifier); ok {
						if rightId, ok := n.Right.(*ast.Identifier); ok {
							if leftId.Value == rightId.Value && !o.mayBeZero(leftId.Value) {
								return midId
							}
						}
					}
				}
			}
		}

		if n.Operator == "*" {
			if leftInfix, ok := n.Left.(*ast.InfixExpression); ok && leftInfix.Operator == "/" {
				if leftId, ok := leftInfix.Left.(*ast.Identifier); ok {
					if midId, ok := leftInfix.Right.(*ast.Identifier); ok {
						if rightId, ok := n.Right.(*ast.Identifier); ok {
							if midId.Value == rightId.Value && !o.mayBeZero(midId.Value) {
								return leftId
							}
						}
					}
				}
			}
			if rightInfix, ok := n.Right.(*ast.InfixExpression); ok && rightInfix.Operator == "/" {
				if leftId, ok := n.Left.(*ast.Identifier); ok {
					if midId, ok := rightInfix.Right.(*ast.Identifier); ok {
						if leftId.Value == midId.Value && !o.mayBeZero(leftId.Value) {
							return rightInfix.Left
						}
					}
				}
			}
		}

		if rightInt, ok := n.Right.(*ast.IntegerLiteral); ok {
			if rightInt.Value == 0 {
				if n.Operator == "*" {
					return &ast.IntegerLiteral{Value: 0}
				}
				if n.Operator == "+" {
					return n.Left
				}
				if n.Operator == "-" {
					return n.Left
				}
			}
			if rightInt.Value == 1 && n.Operator == "*" {
				return n.Left
			}
		}

		if leftInt, ok := n.Left.(*ast.IntegerLiteral); ok {
			if leftInt.Value == 0 && n.Operator == "+" {
				return n.Right
			}
		}

		if folded := o.evaluateConstant(n); folded != nil {
			return folded
		}
		return n

	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "print", "input", "get_time", "len", "push", "pop", "first", "rest", "contains", "negate":
				return n
			}
		}

		for i, arg := range n.Arguments {
			n.Arguments[i] = o.constantFoldExpression(arg)
		}
		return n

	case *ast.IfExpression:
		n.Condition = o.constantFoldExpression(n.Condition)
		if n.Consequence != nil {
			n.Consequence = o.constantFold(n.Consequence).(*ast.BlockStatement)
		}
		for i, elif := range n.Alternative {
			elif.Condition = o.constantFoldExpression(elif.Condition)
			if elif.Consequence != nil {
				elif.Consequence = o.constantFold(elif.Consequence).(*ast.BlockStatement)
			}
			n.Alternative[i] = elif
		}
		if n.Else != nil {
			n.Else = o.constantFold(n.Else).(*ast.BlockStatement)
		}
		return n

	case *ast.ForStatement:
		if folded := o.foldLoopToFormula(n); folded != nil {
			return folded
		}
		if folded := o.foldArithmeticProgression(n); folded != nil {
			return folded
		}

		if n.Initialization != nil {
			init := o.constantFold(n.Initialization)
			if init != nil {
				n.Initialization = init.(*ast.AssignStatement)
			}
		}
		if n.Condition != nil {
			cond := o.constantFoldExpression(n.Condition)
			if cond != nil {
				n.Condition = cond
			}
		}
		if n.Update != nil {
			update := o.constantFold(n.Update)
			if update != nil {
				n.Update = update.(*ast.AssignStatement)
			}
		}
		if n.Body != nil {
			body := o.constantFold(n.Body)
			if body != nil {
				n.Body = body.(*ast.BlockStatement)
			}
		}

		if n.Body != nil && len(n.Body.Statements) == 0 {
			return nil
		}

		return n

	case *ast.WhileStatement:
		if n.Condition != nil {
			cond := o.constantFoldExpression(n.Condition)
			if cond != nil {
				n.Condition = cond
			}
		}
		if n.Body != nil {
			body := o.constantFold(n.Body)
			if body != nil {
				n.Body = body.(*ast.BlockStatement)
			}
		}

		if n.Body != nil && len(n.Body.Statements) == 0 {
			return nil
		}

		return n

	case *ast.PrefixExpression:
		n.Right = o.constantFoldExpression(n.Right)
		if folded := o.evaluateConstant(n); folded != nil {
			return folded
		}
		return n

	default:
		return n
	}
}

func (o *Optimizer) constantFoldExpression(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}
	result := o.constantFold(expr)
	if result == nil {
		return expr
	}
	if exprResult, ok := result.(ast.Expression); ok {
		return exprResult
	}
	return expr
}

func (o *Optimizer) evaluateConstant(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	if call, ok := expr.(*ast.CallExpression); ok {
		if ident, ok := call.Function.(*ast.Identifier); ok {
			if ident.Value == "get_time" || ident.Value == "input" {
				return expr
			}
		}
	}

	if o.hasUserFunctionCalls(expr) {
		return expr
	}

	if o.hasVariables(expr) {
		return expr
	}

	tempEnv := evaluator.NewEnvironment()
	result := evaluator.Eval(expr, tempEnv)

	if result.IsNil() {
		return expr
	}

	return o.valueToLiteral(result)
}

func (o *Optimizer) hasUserFunctionCalls(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "print", "input", "get_time", "len", "push", "pop", "first", "rest", "contains", "negate", "repeat", "reserve":
				return false
			default:
				return true
			}
		}
		return true
	case *ast.InfixExpression:
		return o.hasUserFunctionCalls(n.Left) || o.hasUserFunctionCalls(n.Right)
	case *ast.AssignStatement:
		return o.hasUserFunctionCalls(n.Value)
	case *ast.PrefixExpression:
		return o.hasUserFunctionCalls(n.Right)
	default:
		return false
	}
}

func (o *Optimizer) substituteForBounds(program *ast.Program) {
	env := make(map[string]int64)
	o.substituteBoundsInStatements(program.Statements, env)
}

func (o *Optimizer) substituteBoundsInStatements(statements []ast.Node, env map[string]int64) {
	for _, stmt := range statements {
		o.substituteBoundsInNode(stmt, env)
	}
}

func (o *Optimizer) substituteBoundsInNode(node ast.Node, env map[string]int64) {
	switch n := node.(type) {
	case *ast.AssignStatement:
		if lit, ok := n.Value.(*ast.IntegerLiteral); ok {
			env[n.Name.Value] = lit.Value
		}
		return

	case *ast.ForStatement:
		if n.Condition != nil {
			if infix, ok := n.Condition.(*ast.InfixExpression); ok {
				if ident, ok := infix.Right.(*ast.Identifier); ok {
					if val, exists := env[ident.Value]; exists {
						infix.Right = &ast.IntegerLiteral{Value: val}
					}
				}
			}
		}
		if n.Body != nil {
			for _, s := range n.Body.Statements {
				o.substituteBoundsInNode(s, env)
			}
		}
		return

	case *ast.BlockStatement:
		for _, s := range n.Statements {
			o.substituteBoundsInNode(s, env)
		}
		return

	case *ast.FunctionStatement:
		if n.Body != nil {
			for _, s := range n.Body.Statements {
				o.substituteBoundsInNode(s, env)
			}
		}
		return

	case *ast.Program:
		for _, s := range n.Statements {
			o.substituteBoundsInNode(s, env)
		}
		return
	}
}
