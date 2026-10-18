package value

import (
	"fmt"
	"strings"
)

type ValueType int

const (
	ValNil ValueType = iota
	ValInt
	ValFloat
	ValBool
	ValString
	ValNull
	ValList
	ValMap
	ValFunction
	ValRepeat
)

func (vt ValueType) String() string {
	switch vt {
	case ValNil:
		return "nil"
	case ValInt:
		return "int"
	case ValFloat:
		return "float"
	case ValBool:
		return "bool"
	case ValString:
		return "string"
	case ValNull:
		return "null"
	case ValList:
		return "list"
	case ValMap:
		return "map"
	case ValFunction:
		return "function"
	case ValRepeat:
		return "repeat"
	default:
		return "unknown"
	}
}

type Value struct {
	Type  ValueType
	Int   int64
	Float float64
	Bool  bool
	Str   string
	Obj   interface{}
}

func (v Value) String() string {
	switch v.Type {
	case ValNil:
		return "nil"
	case ValInt:
		return fmt.Sprintf("%d", v.Int)
	case ValFloat:
		return fmt.Sprintf("%g", v.Float)
	case ValBool:
		return fmt.Sprintf("%t", v.Bool)
	case ValString:
		return v.Str
	case ValNull:
		return "null"
	case ValList:
		if list, ok := v.Obj.(*ListObject); ok {
			return list.Inspect()
		}
		return "[...]"
	case ValMap:
		if mapObj, ok := v.Obj.(*MapObject); ok {
			return mapObj.Inspect()
		}
		return "{...}"
	case ValFunction:
		return "function"
	default:
		return "unknown"
	}
}

type ListObject struct {
	Elements []Value
}

func (lo *ListObject) Inspect() string {
	var out strings.Builder
	out.WriteString("[")
	for i, el := range lo.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(el.String())
	}
	out.WriteString("]")
	return out.String()
}

type MapObject struct {
	Elements map[Value]Value
}

func (mo *MapObject) Inspect() string {
	var out strings.Builder
	out.WriteString("{")
	i := 0
	for key, value := range mo.Elements {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(key.String())
		out.WriteString(" :: ")
		out.WriteString(value.String())
		i++
	}
	out.WriteString("}")
	return out.String()
}

type FunctionObject struct {
	Parameters []interface{}
	Body       interface{}
	Env        interface{}
	NumLocals  int
	NativeFn   func([]Value) Value
}

func NewIntValue(v int64) Value {
	return Value{Type: ValInt, Int: v}
}

func NewFloatValue(v float64) Value {
	return Value{Type: ValFloat, Float: v}
}

func NewBoolValue(v bool) Value {
	return Value{Type: ValBool, Bool: v}
}

func NewStringValue(v string) Value {
	return Value{Type: ValString, Str: v}
}

func NewListValue(v *ListObject) Value {
	return Value{Type: ValList, Obj: v}
}

func NewMapValue(v *MapObject) Value {
	return Value{Type: ValMap, Obj: v}
}

func NewNullValue() Value {
	return Value{Type: ValNull}
}

func NewFunctionValue(v *FunctionObject) Value {
	return Value{Type: ValFunction, Obj: v}
}

func NewNilValue() Value {
	return Value{Type: ValNil}
}

func (v Value) IsNil() bool {
	return v.Type == ValNil
}

func (v Value) IsInt() bool {
	return v.Type == ValInt
}

func (v Value) IsFloat() bool {
	return v.Type == ValFloat
}

func (v Value) IsNumber() bool {
	return v.Type == ValInt || v.Type == ValFloat
}

func (v Value) IsBool() bool {
	return v.Type == ValBool
}

func (v Value) IsString() bool {
	return v.Type == ValString
}

func (v Value) IsList() bool {
	return v.Type == ValList
}

func (v Value) IsMap() bool {
	return v.Type == ValMap
}

func Null() Value {
	return Value{Type: ValNull}
}

func (v Value) IsNull() bool {
	return v.Type == ValNull
}

func (v Value) IsFunction() bool {
	return v.Type == ValFunction
}

func (v Value) ToFloat64() (float64, bool) {
	switch v.Type {
	case ValInt:
		return float64(v.Int), true
	case ValFloat:
		return v.Float, true
	default:
		return 0, false
	}
}

func (v Value) ToInt64() (int64, bool) {
	switch v.Type {
	case ValInt:
		return v.Int, true
	case ValFloat:
		return int64(v.Float), true
	default:
		return 0, false
	}
}

type RepeatObject struct {
	Value Value
	Count int64
}

func (ro *RepeatObject) Inspect() string {
	return fmt.Sprintf("repeat(%v, %d)", ro.Value, ro.Count)
}

func (ro *RepeatObject) Get(index int64) Value {
	if index >= 0 && index < ro.Count {
		return ro.Value
	}
	return NewNilValue()
}

func (ro *RepeatObject) Len() int64 {
	return ro.Count
}
