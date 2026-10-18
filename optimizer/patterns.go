package optimizer

import (
	"Tokype/ast"
)

func (o *Optimizer) foldArithmeticProgression(forStmt *ast.ForStatement) ast.Node {
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

	leftIdent, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 {
		return nil
	}

	if leftIdent.Value != assign.Name.Value {
		return nil
	}

	rightIdent, ok2 := infix.Right.(*ast.Identifier)
	if !ok2 {
		return o.foldSimpleArithmetic(forStmt)
	}

	if rightIdent.Value != forStmt.Initialization.Name.Value {
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

	count := endInt.Value - startVal.Value + 1
	if count <= 0 {
		return nil
	}

	sum := count * (startVal.Value + endInt.Value) / 2

	return &ast.AssignStatement{
		Name:  assign.Name,
		Value: &ast.IntegerLiteral{Value: sum},
	}
}

func (o *Optimizer) foldSimpleArithmetic(forStmt *ast.ForStatement) ast.Node {
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

	leftIdent, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 {
		return nil
	}

	if leftIdent.Value != assign.Name.Value {
		return nil
	}

	_, ok2 := infix.Right.(*ast.IntegerLiteral)
	if !ok2 {
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

	count := endInt.Value - startVal.Value + 1
	if count <= 0 {
		return nil
	}

	return &ast.AssignStatement{
		Name:  assign.Name,
		Value: &ast.IntegerLiteral{Value: count},
	}
}

func (o *Optimizer) inlineFunctionCalls(program *ast.Program) *ast.Program {
	if program == nil {
		return program
	}

	newStatements := []ast.Node{}
	functionCalls := make(map[string][]*ast.CallExpression)
	functionNames := make(map[string]bool)

	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Name != nil {
				functionNames[fnStmt.Name.Value] = true
			}
		}
		o.collectCalls(stmt, functionCalls)
	}

	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if calls, exists := functionCalls[fnStmt.Name.Value]; exists && len(calls) > 0 {
				for _, call := range calls {
					inlined := o.inlineFunctionCall(call, fnStmt)
					if inlined != nil {
						newStatements = append(newStatements, inlined)
					}
				}
			} else {
				newStatements = append(newStatements, fnStmt)
			}
		} else {
			if call, ok := stmt.(*ast.CallExpression); ok {
				if ident, ok := call.Function.(*ast.Identifier); ok {
					if functionNames[ident.Value] {
						continue
					}
				}
			}
			newStatements = append(newStatements, stmt)
		}
	}

	program.Statements = newStatements
	return program
}

func (o *Optimizer) collectCalls(node ast.Node, calls map[string][]*ast.CallExpression) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			o.collectCalls(stmt, calls)
		}
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			o.collectCalls(stmt, calls)
		}
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			if _, exists := calls[ident.Value]; !exists {
				calls[ident.Value] = []*ast.CallExpression{}
			}
			calls[ident.Value] = append(calls[ident.Value], n)
		}
	}
}

func (o *Optimizer) inlineFunctionCall(call *ast.CallExpression, fnStmt *ast.FunctionStatement) ast.Node {
	if fnStmt.Body == nil {
		return call
	}

	newBody := &ast.BlockStatement{
		Statements: []ast.Node{},
	}

	for i, param := range fnStmt.Parameters {
		if i < len(call.Arguments) {
			assign := &ast.AssignStatement{
				Name:  &ast.Identifier{Value: param.Value},
				Value: call.Arguments[i],
			}
			newBody.Statements = append(newBody.Statements, assign)
		}
	}

	for _, stmt := range fnStmt.Body.Statements {
		newBody.Statements = append(newBody.Statements, stmt)
	}

	return newBody
}

func (o *Optimizer) removeEmptyFunctions(program *ast.Program) *ast.Program {
	newStatements := []ast.Node{}
	for _, stmt := range program.Statements {
		if fnStmt, ok := stmt.(*ast.FunctionStatement); ok {
			if fnStmt.Body != nil && len(fnStmt.Body.Statements) > 0 {
				newStatements = append(newStatements, stmt)
			}
		} else {
			newStatements = append(newStatements, stmt)
		}
	}
	program.Statements = newStatements
	return program
}
