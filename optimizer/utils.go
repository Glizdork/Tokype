package optimizer

import (
	"Tokype/ast"
	"Tokype/value"
)

func (o *Optimizer) valueToLiteral(v value.Value) ast.Expression {
	switch v.Type {
	case value.ValInt:
		return &ast.IntegerLiteral{Value: v.Int}
	case value.ValFloat:
		return &ast.FloatLiteral{Value: v.Float}
	case value.ValBool:
		return &ast.BooleanLiteral{Value: v.Bool}
	case value.ValString:
		return &ast.StringLiteral{Value: v.Str}
	case value.ValNull:
		return &ast.NullLiteral{}
	default:
		return nil
	}
}

func (o *Optimizer) hasVariables(node ast.Node) bool {
	if node == nil {
		return false
	}

	o.cacheMutex.RLock()
	if val, ok := o.cache[node]; ok {
		o.cacheMutex.RUnlock()
		return val
	}
	o.cacheMutex.RUnlock()

	var result bool
	switch n := node.(type) {
	case *ast.Identifier:
		if _, isFunction := o.functions[n.Value]; isFunction {
			return false
		}
		return true
	case *ast.Program:
		for _, stmt := range n.Statements {
			if o.hasVariables(stmt) {
				result = true
				break
			}
		}
	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			if o.hasVariables(stmt) {
				result = true
				break
			}
		}
	case *ast.AssignStatement:
		result = o.hasVariables(n.Value)
	case *ast.InfixExpression:
		result = o.hasVariables(n.Left) || o.hasVariables(n.Right)
	case *ast.CallExpression:
		if ident, ok := n.Function.(*ast.Identifier); ok {
			switch ident.Value {
			case "print", "input", "get_time", "len", "push", "pop", "first", "rest", "contains", "negate":
				result = false
			}
		}
		if !result {
			for _, arg := range n.Arguments {
				if o.hasVariables(arg) {
					result = true
					break
				}
			}
		}
	case *ast.FunctionStatement:
		result = o.hasVariables(n.Body)
	case *ast.PrefixExpression:
		result = o.hasVariables(n.Right)
	default:
		result = false
	}

	o.cacheMutex.Lock()
	o.cache[node] = result
	o.cacheMutex.Unlock()

	return result
}

func (o *Optimizer) hasSideEffects(funcName string) bool {
	switch funcName {
	case "print", "input", "push", "pop":
		return true
	default:
		return false
	}
}
