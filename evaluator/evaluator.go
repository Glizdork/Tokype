package evaluator

import (
	"Tokype/ast"
	"Tokype/value"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var ProgramStartTime = time.Now()

type Environment struct {
	store map[string]value.Value
	cache map[string]int
	data  []value.Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]value.Value),
		cache: make(map[string]int),
		data:  make([]value.Value, 0, 16),
		outer: nil,
	}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Set(name string, val value.Value) value.Value {
	if id, ok := e.cache[name]; ok {
		e.data[id] = val
		e.store[name] = val
		return val
	}

	id := len(e.data)
	e.cache[name] = id
	e.data = append(e.data, val)
	e.store[name] = val
	return val
}

func (e *Environment) Get(name string) (value.Value, bool) {
	if id, ok := e.cache[name]; ok {
		if id < len(e.data) {
			return e.data[id], true
		}
	}

	val, ok := e.store[name]
	if ok {
		id := len(e.data)
		e.cache[name] = id
		e.data = append(e.data, val)
		return val, true
	}

	if e.outer != nil {
		return e.outer.Get(name)
	}

	return value.NewNilValue(), false
}

type ReturnSignal struct {
	Value value.Value
}

var FunctionStore = make(map[string]*value.FunctionObject)
var Reader = bufio.NewReader(os.Stdin)

func isNumber(v value.Value) bool {
	return v.Type == value.ValInt || v.Type == value.ValFloat
}

func valuesEqual(a, b value.Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case value.ValInt:
		return a.Int == b.Int
	case value.ValFloat:
		return a.Float == b.Float
	case value.ValString:
		return a.Str == b.Str
	case value.ValBool:
		return a.Bool == b.Bool
	case value.ValNull:
		return true
	case value.ValNil:
		return true
	}
	return false
}

func evalNumberOperation(left, right value.Value, operator string, line, col int) value.Value {
	if left.Type == value.ValInt && right.Type == value.ValInt {
		switch operator {
		case "+":
			return value.NewIntValue(left.Int + right.Int)
		case "-":
			return value.NewIntValue(left.Int - right.Int)
		case "*":
			return value.NewIntValue(left.Int * right.Int)
		case "/":
			if right.Int == 0 {
				fmt.Printf("Error: division by zero [line: %d, column: %d]\n", line, col)
				os.Exit(0)
				return value.NewNilValue()
			}
			if left.Int%right.Int == 0 {
				return value.NewIntValue(left.Int / right.Int)
			}
			return value.NewFloatValue(float64(left.Int) / float64(right.Int))
		case "%":
			if right.Int == 0 {
				fmt.Printf("Error: modulo by zero [line: %d, column: %d]\n", line, col)
				os.Exit(0)
				return value.NewNilValue()
			}
			return value.NewIntValue(left.Int % right.Int)
		}
	}

	leftVal, ok1 := left.ToFloat64()
	rightVal, ok2 := right.ToFloat64()

	if !ok1 || !ok2 {
		return value.NewNilValue()
	}

	var result float64
	switch operator {
	case "+":
		result = leftVal + rightVal
	case "-":
		result = leftVal - rightVal
	case "*":
		result = leftVal * rightVal
	case "/":
		if rightVal == 0 {
			fmt.Printf("Error: division by zero [line: %d, column: %d]\n", line, col)
			os.Exit(0)
			return value.NewNilValue()
		}
		result = leftVal / rightVal
	case "%":
		if rightVal == 0 {
			fmt.Printf("Error: modulo by zero [line: %d, column: %d]\n", line, col)
			os.Exit(0)
			return value.NewNilValue()
		}
		result = float64(int64(leftVal) % int64(rightVal))
	default:
		return value.NewNilValue()
	}

	if result == float64(int64(result)) {
		return value.NewIntValue(int64(result))
	}
	return value.NewFloatValue(result)
}

func evalNumberComparison(left, right value.Value, operator string) value.Value {
	leftVal, ok1 := left.ToFloat64()
	rightVal, ok2 := right.ToFloat64()

	if !ok1 || !ok2 {
		return value.NewNilValue()
	}

	switch operator {
	case "<":
		return value.NewBoolValue(leftVal < rightVal)
	case ">":
		return value.NewBoolValue(leftVal > rightVal)
	case "<=":
		return value.NewBoolValue(leftVal <= rightVal)
	case ">=":
		return value.NewBoolValue(leftVal >= rightVal)
	default:
		return value.NewNilValue()
	}
}

func evalPrefixExpression(operator string, right value.Value, line, col int) value.Value {
	switch operator {
	case "-":
		if right.Type == value.ValInt {
			return value.NewIntValue(-right.Int)
		}
		if right.Type == value.ValFloat {
			return value.NewFloatValue(-right.Float)
		}
		fmt.Printf("Error: unknown operator: -%s [line: %d, column: %d]\n", right.Type.String(), line, col)
		os.Exit(0)
		return value.Null()
	case "!":
		if right.Type == value.ValBool {
			return value.NewBoolValue(!right.Bool)
		}
		fmt.Printf("Error: !%s is not supported [line: %d, column: %d]\n", right.Type.String(), line, col)
		os.Exit(1)
		return value.Null()
	default:
		fmt.Printf("Error: unknown operator: %s%s [line: %d, column: %d]\n", operator, right.Type.String(), line, col)
		os.Exit(0)
		return value.Null()
	}
}

func Eval(node ast.Node, env *Environment) (result value.Value) {
	defer func() {
		if r := recover(); r != nil {
			if rtErr, ok := r.(*value.RuntimeError); ok {
				fmt.Printf("Runtime error: %s\n", rtErr.Error())
				result = value.NewNilValue()
				os.Exit(1)
				return
			}
			panic(r)
		}
	}()

	switch node := node.(type) {
	case *ast.Program:
		var result value.Value = value.NewNilValue()
		for _, statement := range node.Statements {
			if _, ok := statement.(*ast.FunctionStatement); ok {
				Eval(statement, env)
			}
		}
		for _, statement := range node.Statements {
			if _, ok := statement.(*ast.FunctionStatement); !ok {
				result = Eval(statement, env)
			}
		}
		return result

	case *ast.BlockStatement:
		var result value.Value = value.NewNilValue()
		for _, statement := range node.Statements {
			result = Eval(statement, env)
		}
		return result

	case *ast.ReturnStatement:
		var val value.Value = value.NewNilValue()
		if node.Value != nil {
			val = Eval(node.Value, env)
		}
		panic(ReturnSignal{Value: val})

	case *ast.DotExpression:
		if leftIdent, ok := node.Left.(*ast.Identifier); ok {
			fullName := leftIdent.Value + "_" + node.Right.Value

			if fn, ok := FunctionStore[fullName]; ok {
				return value.NewFunctionValue(fn)
			}
			if val, ok := env.Get(fullName); ok {
				return val
			}
		}
		return value.NewNilValue()

	case *ast.IntegerLiteral:
		return value.NewIntValue(node.Value)

	case *ast.FloatLiteral:
		return value.NewFloatValue(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if right.IsNil() || right.IsNull() {
			return value.Null()
		}
		return evalPrefixExpression(node.Operator, right, node.Line, node.Column)

	case *ast.StringLiteral:
		return value.NewStringValue(node.Value)

	case *ast.BooleanLiteral:
		return value.NewBoolValue(node.Value)

	case *ast.ListLiteral:
		elements := []value.Value{}
		for _, el := range node.Elements {
			evaluated := Eval(el, env)
			elements = append(elements, evaluated)
		}
		return value.NewListValue(&value.ListObject{Elements: elements})

	case *ast.MapLiteral:
		pairs := make(map[value.Value]value.Value)
		for keyNode, valueNode := range node.Pairs {
			key := Eval(keyNode, env)
			val := Eval(valueNode, env)
			pairs[key] = val
		}
		return value.NewMapValue(&value.MapObject{Elements: pairs})

	case *ast.Identifier:
		val, ok := env.Get(node.Value)
		if !ok {
			if fn, ok := FunctionStore[node.Value]; ok {
				return value.NewFunctionValue(fn)
			}
			return value.NewNilValue()
		}
		return val

	case *ast.NullLiteral:
		return value.Null()

	case *ast.InfixExpression:
		leftVal := Eval(node.Left, env)
		rightVal := Eval(node.Right, env)

		if leftVal.IsNil() || rightVal.IsNil() {
			return value.NewNilValue()
		}

		if leftVal.IsNull() && rightVal.IsNull() {
			switch node.Operator {
			case "==":
				return value.NewBoolValue(true)
			case "?=":
				return value.NewBoolValue(false)
			default:
				return value.NewNullValue()
			}
		}

		if leftVal.IsNull() || rightVal.IsNull() {
			switch node.Operator {
			case "==":
				return value.NewBoolValue(false)
			case "?=":
				return value.NewBoolValue(true)
			default:
				return value.NewNullValue()
			}
		}

		switch node.Operator {
		case "+":
			if leftVal.Type == value.ValString {
				return value.NewStringValue(leftVal.Str + rightVal.String())
			}
			if rightVal.Type == value.ValString {
				return value.NewStringValue(leftVal.String() + rightVal.Str)
			}
			return evalNumberOperation(leftVal, rightVal, "+", node.Line, node.Column)
		case "-":
			if leftVal.Type == value.ValString && rightVal.Type == value.ValString {
				result := leftVal.Str
				for _, ch := range rightVal.Str {
					result = strings.ReplaceAll(result, string(ch), "")
				}
				return value.NewStringValue(result)
			}
			if leftVal.Type == value.ValList && rightVal.Type == value.ValList {
				leftList := leftVal.Obj.(*value.ListObject)
				rightList := rightVal.Obj.(*value.ListObject)
				newElements := []value.Value{}
				for _, el := range leftList.Elements {
					found := false
					for _, rem := range rightList.Elements {
						if valuesEqual(el, rem) {
							found = true
							break
						}
					}
					if !found {
						newElements = append(newElements, el)
					}
				}
				return value.NewListValue(&value.ListObject{Elements: newElements})
			}
			if leftVal.Type == value.ValList && rightVal.Type != value.ValList {
				list, ok := leftVal.Obj.(*value.ListObject)
				if ok {
					newElements := []value.Value{}
					for _, el := range list.Elements {
						if !valuesEqual(el, rightVal) {
							newElements = append(newElements, el)
						}
					}
					return value.NewListValue(&value.ListObject{Elements: newElements})
				}
			}
			return evalNumberOperation(leftVal, rightVal, "-", node.Line, node.Column)
		case "*":
			if leftVal.Type == value.ValString && rightVal.Type == value.ValInt {
				result := ""
				for i := int64(0); i < rightVal.Int; i++ {
					result += leftVal.Str
				}
				return value.NewStringValue(result)
			}
			if leftVal.Type == value.ValList && rightVal.Type == value.ValInt {
				list := leftVal.Obj.(*value.ListObject)
				newElements := []value.Value{}
				for i := int64(0); i < rightVal.Int; i++ {
					newElements = append(newElements, list.Elements...)
				}
				return value.NewListValue(&value.ListObject{Elements: newElements})
			}
			return evalNumberOperation(leftVal, rightVal, "*", node.Line, node.Column)
		case "/":
			return evalNumberOperation(leftVal, rightVal, "/", node.Line, node.Column)
		case "%":
			return evalNumberOperation(leftVal, rightVal, "%", node.Line, node.Column)
		case "==":
			if isNumber(leftVal) && isNumber(rightVal) {
				l, _ := leftVal.ToFloat64()
				r, _ := rightVal.ToFloat64()
				return value.NewBoolValue(l == r)
			}
			return value.NewBoolValue(leftVal == rightVal)
		case "?=":
			if isNumber(leftVal) && isNumber(rightVal) {
				l, _ := leftVal.ToFloat64()
				r, _ := rightVal.ToFloat64()
				return value.NewBoolValue(l != r)
			}
			return value.NewBoolValue(leftVal != rightVal)
		case "<", ">", "<=", ">=":
			return evalNumberComparison(leftVal, rightVal, node.Operator)
		default:
			return value.NewNilValue()
		}

	case *ast.AssignStatement:
		val := Eval(node.Value, env)
		env.Set(node.Name.Value, val)
		return val

	case *ast.FunctionStatement:
		fn := &value.FunctionObject{
			Parameters: make([]interface{}, len(node.Parameters)),
			Body:       node.Body,
			Env:        env,
		}
		for i, p := range node.Parameters {
			fn.Parameters[i] = p.Value
		}
		FunctionStore[node.Name.Value] = fn
		return value.NewFunctionValue(fn)

	case *ast.ForStatement:
		Eval(node.Initialization, env)

		for {
			cond := Eval(node.Condition, env)
			if cond.IsNil() {
				break
			}
			if cond.Type == value.ValBool && !cond.Bool {
				break
			}

			Eval(node.Body, env)
			Eval(node.Update, env)
		}
		return value.NewNilValue()

	case *ast.WhileStatement:
		for {
			cond := Eval(node.Condition, env)
			if cond.IsNil() || cond.IsNull() {
				break
			}
			if cond.Type == value.ValBool && !cond.Bool {
				break
			}
			Eval(node.Body, env)
		}
		return value.NewNilValue()

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}
		index := Eval(node.Index, env)
		if index.IsNil() {
			return value.NewNilValue()
		}

		if left.Type == value.ValRepeat {
			repeatObj, ok := left.Obj.(*value.RepeatObject)
			if !ok {
				fmt.Printf("Error: invalid repeat object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Printf("Error: index must be a number [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			if idx < 0 || idx >= repeatObj.Count {
				fmt.Printf("Error: index %d is out of bounds (length: %d)[line: %d, column: %d]\n", idx, repeatObj.Count, node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			return repeatObj.Value
		}

		if left.Type == value.ValList {
			list, ok := left.Obj.(*value.ListObject)
			if !ok {
				fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Printf("Error: index must be a number [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			if idx < 0 || idx >= int64(len(list.Elements)) {
				fmt.Printf("Error: index %d is out of list bounds (length: %d) [line: %d, column: %d]\n", idx, len(list.Elements), node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			return list.Elements[idx]
		}

		if left.Type == value.ValMap {
			mapObj, ok := left.Obj.(*value.MapObject)
			if !ok {
				fmt.Printf("Error: invalid map object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			val, ok := mapObj.Elements[index]
			if !ok {
				fmt.Printf("Error: key %v not found in the dictionary [line: %d, column: %d]\n", index, node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			return val
		}

		fmt.Printf("Error: indexing is not supported for the type %s [line: %d, column: %d]\n", left.Type.String(), node.Line, node.Column)
		os.Exit(1)
		return value.NewNilValue()

	case *ast.AssignIndexStatement:
		left := Eval(node.Left.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}
		index := Eval(node.Left.Index, env)
		if index.IsNil() {
			return value.NewNilValue()
		}
		val := Eval(node.Value, env)
		if val.IsNil() {
			return value.NewNilValue()
		}

		if left.Type == value.ValRepeat {
			fmt.Printf("Error: cannot assign to repeat object (read-only) [line: %d, column: %d]\n", node.Line, node.Column)
			os.Exit(1)
			return value.NewNilValue()
		}

		if left.Type == value.ValList {
			list, ok := left.Obj.(*value.ListObject)
			if !ok {
				fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			idx, ok := index.ToInt64()
			if !ok {
				fmt.Printf("Error: index must be a number [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			if idx < 0 || idx >= int64(len(list.Elements)) {
				fmt.Printf("Error: index %d out of bounds [line: %d, column: %d]\n", idx, node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			list.Elements[idx] = val
			return val
		}

		if left.Type == value.ValMap {
			mapObj, ok := left.Obj.(*value.MapObject)
			if !ok {
				fmt.Printf("Error: invalid map object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			mapObj.Elements[index] = val
			return val
		}

		fmt.Printf("Error: index assignment not supported for %s [line: %d, column: %d]\n", left.Type.String(), node.Line, node.Column)
		os.Exit(1)
		return value.NewNilValue()

	case *ast.SliceExpression:
		left := Eval(node.Left, env)
		if left.IsNil() {
			return value.NewNilValue()
		}

		if left.Type != value.ValList {
			fmt.Printf("Error: slicing works only with lists; got %s [line: %d, column: %d]\n", left.Type.String(), node.Line, node.Column)
			os.Exit(1)
			return value.NewNilValue()
		}

		list, ok := left.Obj.(*value.ListObject)
		if !ok {
			fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
			os.Exit(1)
			return value.NewNilValue()
		}

		start := 0
		end := len(list.Elements)

		if node.Start != nil {
			startVal := Eval(node.Start, env)
			if startVal.IsNil() {
				return value.NewNilValue()
			}
			s, ok := startVal.ToInt64()
			if !ok {
				fmt.Printf("Error: slice start must be a number [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			start = int(s)
		}

		if node.End != nil {
			endVal := Eval(node.End, env)
			if endVal.IsNil() {
				return value.NewNilValue()
			}
			e, ok := endVal.ToInt64()
			if !ok {
				fmt.Printf("Error: slice end must be a number [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			end = int(e)
		}

		if start < 0 {
			start = 0
		}
		if end > len(list.Elements) {
			end = len(list.Elements)
		}
		if start > end {
			return value.NewListValue(&value.ListObject{Elements: []value.Value{}})
		}

		return value.NewListValue(&value.ListObject{
			Elements: list.Elements[start:end],
		})

	case *ast.IfExpression:
		condition := Eval(node.Condition, env)
		if condition.IsNil() {
			return value.NewNilValue()
		}

		if condition.Type == value.ValBool && condition.Bool {
			return Eval(node.Consequence, env)
		}

		for _, elif := range node.Alternative {
			elifCond := Eval(elif.Condition, env)
			if elifCond.IsNil() {
				continue
			}
			if elifCond.Type == value.ValBool && elifCond.Bool {
				return Eval(elif.Consequence, env)
			}
		}

		if node.Else != nil {
			return Eval(node.Else, env)
		}

		return value.NewNilValue()

	case *ast.CallExpression:
		var funcName string
		var fnObj *value.FunctionObject

		switch fn := node.Function.(type) {
		case *ast.Identifier:
			funcName = fn.Value

			if funcName == "repeat" {
				if len(node.Arguments) != 2 {
					fmt.Printf("Error: repeat() requires 2 arguments (value, count) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				val := Eval(node.Arguments[0], env)
				countVal := Eval(node.Arguments[1], env)

				if val.IsNil() || countVal.IsNil() {
					return value.NewNilValue()
				}

				count, ok := countVal.ToInt64()
				if !ok {
					fmt.Printf("Error: count must be a number [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				if count < 0 || count > 1000000000 {
					fmt.Printf("Error: count %d is too large (max 1,000,000,000) [line: %d, column: %d]\n", count, node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				if count <= 100000 {
					elements := make([]value.Value, count)
					for i := int64(0); i < count; i++ {
						elements[i] = val
					}
					return value.NewListValue(&value.ListObject{Elements: elements})
				}

				return value.Value{
					Type: value.ValRepeat,
					Obj:  &value.RepeatObject{Value: val, Count: count},
				}
			}

			if funcName == "reserve" {
				if len(node.Arguments) != 2 {
					fmt.Printf("Error: reserve() requires 2 arguments (list, capacity) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				listArg := Eval(node.Arguments[0], env)

				var list *value.ListObject
				if listArg.IsNil() || listArg.Type != value.ValList {
					list = &value.ListObject{
						Elements: []value.Value{},
					}
				} else {
					var ok bool
					list, ok = listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
				}

				capacityVal := Eval(node.Arguments[1], env)
				if capacityVal.IsNil() {
					fmt.Printf("Error: capacity is nil [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				capacity, ok := capacityVal.ToInt64()
				if !ok {
					fmt.Printf("Error: capacity must be a number [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				if capacity < 0 {
					fmt.Printf("Error: capacity cannot be negative [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				if capacity > 100000000 {
					fmt.Printf("Error: capacity %d is too large (max 100,000,000) [line: %d, column: %d]\n", capacity, node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				newList := &value.ListObject{
					Elements: make([]value.Value, 0, capacity),
				}

				if list != nil {
					newList.Elements = append(newList.Elements, list.Elements...)
				}

				return value.NewListValue(newList)
			}

			if funcName == "print" {
				for i, arg := range node.Arguments {
					val := Eval(arg, env)
					if i > 0 {
						fmt.Print(" ")
					}
					if !val.IsNil() {
						fmt.Print(val.String())
					}
				}
				fmt.Println()
				return value.NewNilValue()
			}

			if funcName == "input" {
				if len(node.Arguments) > 0 {
					prompt := Eval(node.Arguments[0], env)
					if !prompt.IsNil() {
						fmt.Print(prompt.String())
					}
				}
				input, _ := Reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if num, err := strconv.ParseFloat(input, 64); err == nil {
					if num == float64(int64(num)) {
						return value.NewIntValue(int64(num))
					}
					return value.NewFloatValue(num)
				}
				if num, err := strconv.ParseInt(input, 10, 64); err == nil {
					return value.NewIntValue(num)
				}
				return value.NewStringValue(input)
			}

			if funcName == "negate" {
				if len(node.Arguments) != 1 {
					fmt.Printf("Error: negate() requires exactly 1 argument [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				arg := Eval(node.Arguments[0], env)
				if arg.IsNil() {
					return value.NewNilValue()
				}
				if arg.Type == value.ValInt {
					return value.NewIntValue(-arg.Int)
				}
				if arg.Type == value.ValFloat {
					return value.NewFloatValue(-arg.Float)
				}
				fmt.Printf("Error: negate() works only with numbers, got %s [line: %d, column: %d]\n", arg.Type.String(), node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "len" {
				if len(node.Arguments) != 1 {
					fmt.Printf("Error: len requires 1 argument [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				arg := Eval(node.Arguments[0], env)
				if arg.IsNil() {
					return value.NewNilValue()
				}
				switch arg.Type {
				case value.ValList:
					list, ok := arg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					return value.NewIntValue(int64(len(list.Elements)))
				case value.ValString:
					return value.NewIntValue(int64(len(arg.Str)))
				default:
					fmt.Printf("Error: len() is not supported for the type %s [line: %d, column: %d]\n", arg.Type.String(), node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
			}

			if funcName == "push" {
				if len(node.Arguments) != 2 {
					fmt.Printf("Error: push() requires 2 arguments (a list and an element) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				elemArg := Eval(node.Arguments[1], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					list.Elements = append(list.Elements, elemArg)
					return listArg
				}
				fmt.Printf("Error: push() works only with lists [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "pop" {
				if len(node.Arguments) != 1 {
					fmt.Printf("Error: pop() requires 1 argument (a list) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						fmt.Printf("Error: pop() on an empty list [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					newList := &value.ListObject{
						Elements: list.Elements[:len(list.Elements)-1],
					}
					return value.NewListValue(newList)
				}
				fmt.Printf("Error: pop() works only with lists [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "first" {
				if len(node.Arguments) != 1 {
					fmt.Printf("Error: first() requires 1 argument (a list) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						return value.NewNilValue()
					}
					return list.Elements[0]
				}
				fmt.Printf("Error: first() works only with lists [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "rest" {
				if len(node.Arguments) != 1 {
					fmt.Printf("Error: rest() requires 1 argument (a list) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					if len(list.Elements) == 0 {
						return value.NewListValue(&value.ListObject{Elements: []value.Value{}})
					}
					return value.NewListValue(&value.ListObject{Elements: list.Elements[1:]})
				}
				fmt.Printf("Error: rest() works only with lists [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "contains" {
				if len(node.Arguments) != 2 {
					fmt.Printf("Error: contains() requires 2 arguments (list and element) [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}
				listArg := Eval(node.Arguments[0], env)
				elemArg := Eval(node.Arguments[1], env)
				if listArg.Type == value.ValList {
					list, ok := listArg.Obj.(*value.ListObject)
					if !ok {
						fmt.Printf("Error: invalid list object [line: %d, column: %d]\n", node.Line, node.Column)
						os.Exit(1)
						return value.NewNilValue()
					}
					for _, el := range list.Elements {
						if el == elemArg {
							return value.NewBoolValue(true)
						}
					}
					return value.NewBoolValue(false)
				}
				fmt.Printf("Error: contains() works only with lists [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

			if funcName == "get_time" {
				if len(node.Arguments) != 0 {
					fmt.Printf("Error: arguments cannot be passed [line: %d, column: %d]\n", node.Line, node.Column)
					os.Exit(1)
					return value.NewNilValue()
				}

				duration := time.Since(ProgramStartTime)

				seconds := duration.Seconds()

				return value.NewFloatValue(seconds)
			}

			var ok bool
			fnObj, ok = FunctionStore[funcName]
			if !ok {
				fmt.Printf("Error: function '%s' not found [line: %d, column: %d]\n", funcName, node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}

		default:
			fnObjVal := Eval(node.Function, env)
			if fnObjVal.IsNil() {
				fmt.Printf("Error: cannot call non-function expression [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			if fnObjVal.Type != value.ValFunction {
				fmt.Printf("Error: cannot call non-function type %s [line: %d, column: %d]\n", fnObjVal.Type.String(), node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
			var ok bool
			fnObj, ok = fnObjVal.Obj.(*value.FunctionObject)
			if !ok {
				fmt.Printf("Error: invalid function object [line: %d, column: %d]\n", node.Line, node.Column)
				os.Exit(1)
				return value.NewNilValue()
			}
		}

		if fnObj == nil {
			fmt.Printf("Error: function object is nil [line: %d, column: %d]\n", node.Line, node.Column)
			os.Exit(1)
			return value.NewNilValue()
		}

		if len(node.Arguments) != len(fnObj.Parameters) {
			fmt.Printf("Error: function expects %d arguments, got %d [line: %d, column: %d]\n", len(fnObj.Parameters), len(node.Arguments), node.Line, node.Column)
			os.Exit(1)
			return value.NewNilValue()
		}

		args := []value.Value{}
		for _, arg := range node.Arguments {
			args = append(args, Eval(arg, env))
		}

		extendedEnv := NewEnclosedEnvironment(fnObj.Env.(*Environment))

		for paramIdx, param := range fnObj.Parameters {
			if paramIdx < len(args) {
				if paramStr, ok := param.(string); ok {
					extendedEnv.Set(paramStr, args[paramIdx])
				}
			}
		}

		var bodyResult value.Value
		func() {
			defer func() {
				if r := recover(); r != nil {
					if signal, ok := r.(ReturnSignal); ok {
						bodyResult = signal.Value
						return
					}
					panic(r)
				}
			}()
			bodyResult = Eval(fnObj.Body.(*ast.BlockStatement), extendedEnv)
		}()
		return bodyResult
	}

	return value.NewNilValue()
}

var stringPool = make(map[string]string)

func Intern(s string) string {
	if cached, ok := stringPool[s]; ok {
		return cached
	}
	copied := string([]byte(s))
	stringPool[s] = copied
	return copied
}
