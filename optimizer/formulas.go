package optimizer

import "Tokype/ast"

func (o *Optimizer) foldLoopToFormula(forStmt *ast.ForStatement) ast.Node {
	if forStmt.Body == nil || len(forStmt.Body.Statements) != 1 {
		return nil
	}

	assign, ok := forStmt.Body.Statements[0].(*ast.AssignStatement)
	if !ok {
		return nil
	}

	infix, ok := assign.Value.(*ast.InfixExpression)
	if !ok {
		return nil
	}

	startVal, ok := forStmt.Initialization.Value.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	endVal, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return nil
	}

	endInt, ok := endVal.Right.(*ast.IntegerLiteral)
	if !ok {
		return nil
	}

	step := int64(1)
	if forStmt.Update != nil {
		if updateInfix, ok := forStmt.Update.Value.(*ast.InfixExpression); ok {
			if intLit, ok := updateInfix.Right.(*ast.IntegerLiteral); ok {
				step = intLit.Value
			}
		}
	}

	if step == 0 {
		return nil
	}

	a := startVal.Value
	b := endInt.Value

	if (step > 0 && a > b) || (step < 0 && a < b) {
		return nil
	}

	var n int64
	if step > 0 {
		n = (b-a)/step + 1
	} else {
		n = (a-b)/(-step) + 1
	}

	if n <= 0 {
		return nil
	}

	leftIdent, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 || leftIdent.Value != assign.Name.Value {
		return nil
	}

	loopVar := forStmt.Initialization.Name.Value

	switch infix.Operator {
	case "+":
		return o.foldSumFormulaWithStep(assign, infix.Right, loopVar, a, b, n, step)
	case "*":
		return o.foldProductFormula(assign, infix.Right, loopVar, a, b, n)
	}

	return nil
}

func (o *Optimizer) foldSumFormulaWithStep(assign *ast.AssignStatement, right ast.Expression, loopVar string, a, _, n, step int64) ast.Node {
	switch r := right.(type) {
	case *ast.Identifier:
		if r.Value != loopVar {
			return nil
		}
		last := a + (n-1)*step
		sum := n * (a + last) / 2
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "+",
				Right:    &ast.IntegerLiteral{Value: sum},
			},
		}

	case *ast.InfixExpression:
		if r.Operator == "*" {
			leftId, ok1 := r.Left.(*ast.Identifier)
			rightId, ok2 := r.Right.(*ast.Identifier)

			if ok1 && ok2 && leftId.Value == loopVar && rightId.Value == loopVar {
				if step == 1 {
					sum := n * (n + 1) * (2*n + 1) / 6
					return &ast.AssignStatement{
						Name: assign.Name,
						Value: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: assign.Name.Value},
							Operator: "+",
							Right:    &ast.IntegerLiteral{Value: sum},
						},
					}
				}
				return nil
			}

			if ok1 && leftId.Value == loopVar {
				if intLit, ok := r.Right.(*ast.IntegerLiteral); ok {
					last := a + (n-1)*step
					sum := intLit.Value * n * (a + last) / 2
					return &ast.AssignStatement{
						Name: assign.Name,
						Value: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: assign.Name.Value},
							Operator: "+",
							Right:    &ast.IntegerLiteral{Value: sum},
						},
					}
				}
			}
			if ok2 && rightId.Value == loopVar {
				if intLit, ok := r.Left.(*ast.IntegerLiteral); ok {
					last := a + (n-1)*step
					sum := intLit.Value * n * (a + last) / 2
					return &ast.AssignStatement{
						Name: assign.Name,
						Value: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: assign.Name.Value},
							Operator: "+",
							Right:    &ast.IntegerLiteral{Value: sum},
						},
					}
				}
			}
		}

	case *ast.IntegerLiteral:
		sum := r.Value * n
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "+",
				Right:    &ast.IntegerLiteral{Value: sum},
			},
		}
	}

	return nil
}

func (o *Optimizer) foldProductFormula(assign *ast.AssignStatement, right ast.Expression, loopVar string, a, b, n int64) ast.Node {
	switch r := right.(type) {
	case *ast.Identifier:
		if r.Value != loopVar {
			return nil
		}
		if a != 1 {
			return nil
		}
		result := int64(1)
		for i := a; i <= b; i++ {
			result *= i
		}
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "*",
				Right:    &ast.IntegerLiteral{Value: result},
			},
		}

	case *ast.IntegerLiteral:
		if a != 1 {
			return nil
		}
		result := int64(1)
		for i := int64(1); i <= n; i++ {
			result *= r.Value
		}
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "*",
				Right:    &ast.IntegerLiteral{Value: result},
			},
		}
	}

	return nil
}

func (o *Optimizer) foldDynamicLoops(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.Program:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.foldDynamicLoops(stmt)
		}
		return n

	case *ast.BlockStatement:
		for i, stmt := range n.Statements {
			n.Statements[i] = o.foldDynamicLoops(stmt)
		}
		return n

	case *ast.FunctionStatement:
		if n.Body != nil {
			n.Body = o.foldDynamicLoops(n.Body).(*ast.BlockStatement)
		}
		return n

	case *ast.ForStatement:
		if n.Body != nil {
			n.Body = o.foldDynamicLoops(n.Body).(*ast.BlockStatement)
		}
		if folded := o.tryDynamicFormula(n); folded != nil {
			return folded
		}
		return n

	default:
		return n
	}
}

func (o *Optimizer) tryDynamicFormula(forStmt *ast.ForStatement) ast.Node {
	if forStmt.Body == nil || len(forStmt.Body.Statements) != 1 {
		return nil
	}

	assign, ok := forStmt.Body.Statements[0].(*ast.AssignStatement)
	if !ok {
		return nil
	}

	infix, ok := assign.Value.(*ast.InfixExpression)
	if !ok || infix.Operator != "+" {
		return nil
	}

	leftIdent, ok := infix.Left.(*ast.Identifier)
	if !ok || leftIdent.Value != assign.Name.Value {
		return nil
	}

	startVal, ok := forStmt.Initialization.Value.(*ast.IntegerLiteral)
	if !ok || startVal.Value != 1 {
		return nil
	}

	infixCond, ok := forStmt.Condition.(*ast.InfixExpression)
	if !ok {
		return nil
	}

	if _, isConst := infixCond.Right.(*ast.IntegerLiteral); isConst {
		return nil
	}

	n := infixCond.Right
	loopVar := forStmt.Initialization.Name.Value

	return o.buildDynamicFormula(assign, infix.Right, loopVar, n)
}

func (o *Optimizer) buildDynamicFormula(assign *ast.AssignStatement, right ast.Expression, loopVar string, n ast.Expression) ast.Node {
	switch r := right.(type) {
	case *ast.Identifier:
		if r.Value != loopVar {
			return nil
		}
		formula := divExpr(
			mulExpr(n, addExpr(n, intLit(1))),
			intLit(2),
		)
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "+",
				Right:    formula,
			},
		}

	case *ast.InfixExpression:
		if r.Operator == "*" {
			leftId, ok1 := r.Left.(*ast.Identifier)
			rightId, ok2 := r.Right.(*ast.Identifier)

			if ok1 && ok2 && leftId.Value == loopVar && rightId.Value == loopVar {
				formula := divExpr(
					mulExpr(
						mulExpr(n, addExpr(n, intLit(1))),
						addExpr(mulExpr(intLit(2), n), intLit(1)),
					),
					intLit(6),
				)
				return &ast.AssignStatement{
					Name: assign.Name,
					Value: &ast.InfixExpression{
						Left:     &ast.Identifier{Value: assign.Name.Value},
						Operator: "+",
						Right:    formula,
					},
				}
			}

			if ok1 && leftId.Value == loopVar {
				if k, ok := r.Right.(*ast.IntegerLiteral); ok {
					formula := divExpr(
						mulExpr(intLit(k.Value), mulExpr(n, addExpr(n, intLit(1)))),
						intLit(2),
					)
					return &ast.AssignStatement{
						Name: assign.Name,
						Value: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: assign.Name.Value},
							Operator: "+",
							Right:    formula,
						},
					}
				}
			}
			if ok2 && rightId.Value == loopVar {
				if k, ok := r.Left.(*ast.IntegerLiteral); ok {
					formula := divExpr(
						mulExpr(intLit(k.Value), mulExpr(n, addExpr(n, intLit(1)))),
						intLit(2),
					)
					return &ast.AssignStatement{
						Name: assign.Name,
						Value: &ast.InfixExpression{
							Left:     &ast.Identifier{Value: assign.Name.Value},
							Operator: "+",
							Right:    formula,
						},
					}
				}
			}

			if inner, ok := r.Left.(*ast.InfixExpression); ok && inner.Operator == "*" {
				if leftId, ok := inner.Left.(*ast.Identifier); ok && leftId.Value == loopVar {
					if midId, ok := inner.Right.(*ast.Identifier); ok && midId.Value == loopVar {
						if rightId, ok := r.Right.(*ast.Identifier); ok && rightId.Value == loopVar {
							s := divExpr(
								mulExpr(n, addExpr(n, intLit(1))),
								intLit(2),
							)
							formula := mulExpr(s, s)
							return &ast.AssignStatement{
								Name: assign.Name,
								Value: &ast.InfixExpression{
									Left:     &ast.Identifier{Value: assign.Name.Value},
									Operator: "+",
									Right:    formula,
								},
							}
						}
					}
				}
			}
		}

	case *ast.IntegerLiteral:
		formula := mulExpr(intLit(r.Value), n)
		return &ast.AssignStatement{
			Name: assign.Name,
			Value: &ast.InfixExpression{
				Left:     &ast.Identifier{Value: assign.Name.Value},
				Operator: "+",
				Right:    formula,
			},
		}
	}

	return nil
}

func intLit(v int64) ast.Expression {
	return &ast.IntegerLiteral{Value: v}
}

func mulExpr(left, right ast.Expression) ast.Expression {
	return &ast.InfixExpression{Left: left, Operator: "*", Right: right}
}

func addExpr(left, right ast.Expression) ast.Expression {
	return &ast.InfixExpression{Left: left, Operator: "+", Right: right}
}

func divExpr(left, right ast.Expression) ast.Expression {
	return &ast.InfixExpression{Left: left, Operator: "/", Right: right}
}
