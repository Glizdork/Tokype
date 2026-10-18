package optimizer

import (
	"Tokype/ast"
	"Tokype/evaluator"
	"Tokype/value"
)

func (o *Optimizer) optimizeControlFlow(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.optimizeControlFlow(stmt)
		}
		return n

	case *ast.BlockStatement:
		newStatements := []ast.Node{}
		for _, stmt := range n.Statements {
			optimized := o.optimizeControlFlow(stmt)
			if optimized != nil {
				newStatements = append(newStatements, optimized)
			}
		}
		n.Statements = newStatements
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.optimizeControlFlow(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.IfExpression:
		if n.Consequence != nil {
			n.Consequence = o.optimizeControlFlow(n.Consequence).(*ast.BlockStatement)
		}
		for i, elif := range n.Alternative {
			elif.Consequence = o.optimizeControlFlow(elif.Consequence).(*ast.BlockStatement)
			n.Alternative[i] = elif
		}
		if n.Else != nil {
			n.Else = o.optimizeControlFlow(n.Else).(*ast.BlockStatement)
		}
		return o.simplifyIf(n)

	case *ast.ForStatement:
		if n.Body != nil {
			n.Body = o.optimizeControlFlow(n.Body).(*ast.BlockStatement)
		}
		return o.simplifyFor(n)

	case *ast.WhileStatement:
		if n.Body != nil {
			n.Body = o.optimizeControlFlow(n.Body).(*ast.BlockStatement)
		}
		return o.simplifyWhile(n)

	default:
		return n
	}
}

func (o *Optimizer) simplifyIf(ifExpr *ast.IfExpression) ast.Node {
	if o.hasUserFunctionCalls(ifExpr.Condition) {
		return ifExpr
	}

	tempEnv := evaluator.NewEnvironment()
	result := evaluator.Eval(ifExpr.Condition, tempEnv)

	if result.Type == value.ValBool && !o.hasVariables(ifExpr.Condition) {
		if result.Bool {
			if ifExpr.Consequence != nil {
				return ifExpr.Consequence
			}
			return &ast.BlockStatement{Statements: []ast.Node{}}
		} else {
			for _, elif := range ifExpr.Alternative {
				if o.hasUserFunctionCalls(elif.Condition) {
					return ifExpr
				}
				elifResult := evaluator.Eval(elif.Condition, tempEnv)
				if elifResult.Type == value.ValBool && elifResult.Bool && !o.hasVariables(elif.Condition) {
					if elif.Consequence != nil {
						return elif.Consequence
					}
				}
			}
			if ifExpr.Else != nil {
				return ifExpr.Else
			}
			return &ast.BlockStatement{Statements: []ast.Node{}}
		}
	}

	return ifExpr
}

func (o *Optimizer) simplifyFor(forStmt *ast.ForStatement) ast.Node {
	condition, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return forStmt
	}

	if o.hasVariables(forStmt.Initialization.Value) || o.hasVariables(condition.Right) {
		return forStmt
	}

	if len(forStmt.Body.Statements) == 1 {
		if assign, ok := forStmt.Body.Statements[0].(*ast.AssignStatement); ok {
			if forStmt.Initialization.Name.Value == assign.Name.Value {
				env := evaluator.NewEnvironment()
				env.Set(forStmt.Initialization.Name.Value, evaluator.Eval(forStmt.Initialization.Value, env))
				evaluator.Eval(assign, env)

				if val, ok := env.Get(forStmt.Initialization.Name.Value); ok {
					newInit := &ast.AssignStatement{
						Name:  &ast.Identifier{Value: forStmt.Initialization.Name.Value},
						Value: o.valueToLiteral(val),
					}

					newCondition := &ast.InfixExpression{
						Left:     &ast.Identifier{Value: forStmt.Initialization.Name.Value},
						Operator: "<=",
						Right:    condition.Right,
					}

					return &ast.ForStatement{
						Initialization: newInit,
						Condition:      newCondition,
						Update:         forStmt.Update,
						Body:           forStmt.Body,
					}
				}
			}
		}
	}

	return forStmt
}

func (o *Optimizer) simplifyWhile(whileStmt *ast.WhileStatement) ast.Node {
	if o.hasUserFunctionCalls(whileStmt.Condition) {
		return whileStmt
	}

	tempEnv := evaluator.NewEnvironment()
	result := evaluator.Eval(whileStmt.Condition, tempEnv)

	if result.Type == value.ValBool && !o.hasVariables(whileStmt.Condition) {
		if !result.Bool {
			return &ast.BlockStatement{Statements: []ast.Node{}}
		}
	}

	return whileStmt
}
