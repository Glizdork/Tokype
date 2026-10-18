package ast

import (
	"bytes"
	"fmt"
	"strings"
)

type Node interface {
	String() string
}

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type Program struct {
	Statements []Node
}

func (p *Program) String() string {
	var out bytes.Buffer
	for i, s := range p.Statements {
		if s != nil {
			out.WriteString(s.String())
			if i < len(p.Statements)-1 {
				out.WriteString("\n")
			}
		}
	}
	return out.String()
}

type Identifier struct {
	Value  string
	Line   int
	Column int
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }

type IntegerLiteral struct {
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) String() string  { return fmt.Sprintf("%d", il.Value) }

type FloatLiteral struct {
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}
func (fl *FloatLiteral) String() string  { return fmt.Sprintf("%g", fl.Value) }

type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string  { return "\"" + sl.Value + "\"" }

type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "true"
	}
	return "false"
}

type NullLiteral struct{}

func (nl *NullLiteral) expressionNode() {}
func (nl *NullLiteral) String() string  { return "null" }

type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
	Line     int
	Column   int
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) String() string {
	if ie.Left == nil || ie.Right == nil {
		return "(nil)"
	}
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

type AssignStatement struct {
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode() {}
func (as *AssignStatement) String() string {
	if as.Name == nil {
		return "nil"
	}
	if as.Value == nil {
		return as.Name.String() + " = nil"
	}
	return as.Name.String() + " = " + as.Value.String()
}

type BlockStatement struct {
	Statements []Node
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for i, s := range bs.Statements {
		if s != nil {
			out.WriteString(s.String())
			if i < len(bs.Statements)-1 {
				out.WriteString("\n")
			}
		}
	}
	return out.String()
}

type FunctionStatement struct {
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode() {}
func (fs *FunctionStatement) String() string {
	var params bytes.Buffer
	for i, p := range fs.Parameters {
		if p != nil {
			params.WriteString(p.String())
		}
		if i < len(fs.Parameters)-1 {
			params.WriteString(", ")
		}
	}
	return "funct " + fs.Name.String() + "(" + params.String() + "):\n" + indent(fs.Body.String(), 4) + "\nend"
}

type CallExpression struct {
	Function  Expression
	Arguments []Expression
	Line      int
	Column    int
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) String() string {
	if ce.Function == nil {
		return "nil()"
	}
	var args []string
	for _, a := range ce.Arguments {
		if a != nil {
			args = append(args, a.String())
		}
	}
	return ce.Function.String() + "(" + strings.Join(args, ", ") + ")"
}

type ListLiteral struct {
	Elements []Expression
}

func (ll *ListLiteral) expressionNode() {}
func (ll *ListLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range ll.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type MapLiteral struct {
	Pairs map[Expression]Expression
}

func (ml *MapLiteral) expressionNode() {}
func (ml *MapLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, val := range ml.Pairs {
		pairs = append(pairs, key.String()+" :: "+val.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type IndexExpression struct {
	Left   Expression
	Index  Expression
	Line   int
	Column int
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + ".[" + ie.Index.String() + "])"
}

type AssignIndexStatement struct {
	Left   *IndexExpression
	Value  Expression
	Line   int
	Column int
}

func (ais *AssignIndexStatement) statementNode() {}
func (ais *AssignIndexStatement) String() string {
	return ais.Left.String() + " = " + ais.Value.String()
}

type SliceExpression struct {
	Left   Expression
	Start  Expression
	End    Expression
	Line   int
	Column int
}

func (se *SliceExpression) expressionNode() {}
func (se *SliceExpression) String() string {
	out := se.Left.String() + ".["
	if se.Start != nil {
		out += se.Start.String()
	}
	out += ".."
	if se.End != nil {
		out += se.End.String()
	}
	out += "]"
	return out
}

type DotExpression struct {
	Left  Expression
	Right *Identifier
}

func (de *DotExpression) expressionNode() {}
func (de *DotExpression) String() string {
	return de.Left.String() + "." + de.Right.String()
}

type IfExpression struct {
	Condition   Expression
	Consequence *BlockStatement
	Alternative []*ElseIfExpression
	Else        *BlockStatement
}

func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString(":\n")
	if ie.Consequence != nil {
		out.WriteString(indent(ie.Consequence.String(), 4))
	}
	for _, alt := range ie.Alternative {
		if alt != nil {
			out.WriteString("\n")
			out.WriteString(alt.String())
		}
	}
	if ie.Else != nil {
		out.WriteString("\nelse:\n")
		out.WriteString(indent(ie.Else.String(), 4))
	}
	out.WriteString("\nend")
	return out.String()
}

func (ie *IfExpression) expressionNode() {}

type ElseIfExpression struct {
	Condition   Expression
	Consequence *BlockStatement
}

func (ei *ElseIfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("elif ")
	out.WriteString(ei.Condition.String())
	out.WriteString(":\n")
	if ei.Consequence != nil {
		out.WriteString(indent(ei.Consequence.String(), 4))
	}
	return out.String()
}

func (ei *ElseIfExpression) expressionNode() {}

type ForStatement struct {
	Initialization *AssignStatement
	Condition      Expression
	Update         *AssignStatement
	Body           *BlockStatement
	Step           Expression
}

func (fs *ForStatement) statementNode()  {}
func (fs *ForStatement) expressionNode() {}
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	if fs.Initialization != nil {
		out.WriteString(fs.Initialization.String())
	}
	if fs.Condition != nil {
		if infix, ok := fs.Condition.(*InfixExpression); ok {
			out.WriteString(", ")
			out.WriteString(infix.Right.String())
		}
	}
	out.WriteString(":\n")
	if fs.Body != nil {
		out.WriteString(indent(fs.Body.String(), 4))
	}
	out.WriteString("\nend")
	return out.String()
}

type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) expressionNode() {}
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString(":\n")
	if ws.Body != nil {
		out.WriteString(indent(ws.Body.String(), 4))
	}
	out.WriteString("\nend")
	return out.String()
}

type ReturnStatement struct {
	Value Expression
}

func (rs *ReturnStatement) statementNode() {}
func (rs *ReturnStatement) String() string {
	if rs.Value != nil {
		return "return " + rs.Value.String()
	}
	return "return"
}

type IncludeStatement struct {
	Path      string
	Namespace string
}

func (is *IncludeStatement) statementNode() {}
func (is *IncludeStatement) String() string {
	if is.Namespace != "" {
		return "include \"" + is.Path + "\" as " + is.Namespace
	}
	return "include \"" + is.Path + "\""
}

type PrefixExpression struct {
	Operator string
	Right    Expression
	Line     int
	Column   int
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

func indent(text string, spaces int) string {
	if text == "" {
		return ""
	}
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
