package ast

import (
	"bytes"
	"strings"
)

type Node interface {
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// ImportStatement: #import<Name> or #import<Name1, Name2>
type ImportStatement struct {
	Modules []string
}

func (is *ImportStatement) statementNode() {}
func (is *ImportStatement) String() string {
	return "#import<" + strings.Join(is.Modules, ", ") + ">"
}

// ImportFileStatement: #importf path | name $var
type ImportFileStatement struct {
	Path     Expression
	Name     Expression
	VarName  string
}

func (ifs *ImportFileStatement) statementNode() {}
func (ifs *ImportFileStatement) String() string {
	return "#importf " + ifs.Path.String() + " | " + ifs.Name.String() + " $" + ifs.VarName
}

// PrintStatement: #print=>type&expression
type PrintStatement struct {
	TypeHint string
	Value    Expression
}

func (ps *PrintStatement) statementNode() {}
func (ps *PrintStatement) String() string {
	if ps.TypeHint != "" {
		return "#print=>" + ps.TypeHint + "&" + ps.Value.String()
	}
	return "#print=>" + ps.Value.String()
}

// VariableStatement: name=>type&value
type VariableStatement struct {
	Name     string
	TypeHint string
	Value    Expression
}

func (vs *VariableStatement) statementNode() {}
func (vs *VariableStatement) String() string {
	if vs.TypeHint != "" {
		return vs.Name + "=>" + vs.TypeHint + "&" + vs.Value.String()
	}
	return vs.Name + "=>" + vs.Value.String()
}

// WhileStatement: #while TRUE=count and.if:condition:
type WhileStatement struct {
	Count     Expression
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode() {}
func (ws *WhileStatement) String() string {
	return "#while " + ws.Count.String() + ": " + ws.Body.String()
}

// ForStatement: #for i in iterable:
type ForStatement struct {
	Iterator string
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode() {}
func (fs *ForStatement) String() string {
	return "#for " + fs.Iterator + " in " + fs.Iterable.String() + ": " + fs.Body.String()
}

// IfStatement: #if condition: body #else: body
type IfStatement struct {
	Condition Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) String() string {
	result := "#if " + is.Condition.String() + ": " + is.Consequence.String()
	if is.Alternative != nil {
		result += " #else: " + is.Alternative.String()
	}
	return result
}

// BlockStatement: { statements }
type BlockStatement struct {
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("{")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString("}")
	return out.String()
}

// FunctionDefinition: #def name(): { body }
type FunctionDefinition struct {
	Name       string
	Parameters []string
	Body       *BlockStatement
}

func (fd *FunctionDefinition) statementNode() {}
func (fd *FunctionDefinition) String() string {
	return "#def " + fd.Name + "(): " + fd.Body.String()
}

// BreakStatement: #break or #break=1 or #break=2
type BreakStatement struct {
	Level int
}

func (bs *BreakStatement) statementNode() {}
func (bs *BreakStatement) String() string {
	if bs.Level > 0 {
		return "#break=" + string(rune('0'+bs.Level))
	}
	return "#break"
}

// ReturnValue: return value from function
type ReturnValue struct {
	Value Expression
}

func (rv *ReturnValue) statementNode() {}
func (rv *ReturnValue) String() string { return "return " + rv.Value.String() }

// ExpressionStatement: expression
type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}
func (es *ExpressionStatement) String() string {
	return es.Expression.String()
}

// Identifier: variable name
type Identifier struct {
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string { return i.Value }

// StringLiteral: "string"
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) String() string { return "\"" + sl.Value + "\"" }

// NumberLiteral: 123 or 3.14
type NumberLiteral struct {
	Value float64
	Raw   string
}

func (nl *NumberLiteral) expressionNode() {}
func (nl *NumberLiteral) String() string { return nl.Raw }

// BooleanLiteral: TRUE or FALSE
type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) String() string {
	if bl.Value {
		return "TRUE"
	}
	return "FALSE"
}

// ListLiteral: ["1", "2"] or str["1", "2"]
type ListLiteral struct {
	TypeHint string
	Elements []Expression
}

func (ll *ListLiteral) expressionNode() {}
func (ll *ListLiteral) String() string {
	elements := []string{}
	for _, e := range ll.Elements {
		elements = append(elements, e.String())
	}
	if ll.TypeHint != "" {
		return ll.TypeHint + "[" + strings.Join(elements, ", ") + "]"
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// DictLiteral: {"key": "value"}
type DictLiteral struct {
	Pairs map[Expression]Expression
}

func (dl *DictLiteral) expressionNode() {}
func (dl *DictLiteral) String() string {
	pairs := []string{}
	for k, v := range dl.Pairs {
		pairs = append(pairs, k.String()+":"+v.String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// PrefixExpression: -5 or !true
type PrefixExpression struct {
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression: 1 + 2, a == b
type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// CallExpression: function(args)
type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) String() string {
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return ce.Function.String() + "(" + strings.Join(args, ", ") + ")"
}

// MethodCallExpression: object.method(args)
type MethodCallExpression struct {
	Object    Expression
	Method    string
	Arguments []Expression
}

func (mce *MethodCallExpression) expressionNode() {}
func (mce *MethodCallExpression) String() string {
	args := []string{}
	for _, a := range mce.Arguments {
		args = append(args, a.String())
	}
	return mce.Object.String() + "." + mce.Method + "(" + strings.Join(args, ", ") + ")"
}

// IndexExpression: array[index]
type IndexExpression struct {
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

// AssignmentExpression: var = value
type AssignmentExpression struct {
	Name  string
	Value Expression
}

func (ae *AssignmentExpression) expressionNode() {}
func (ae *AssignmentExpression) String() string {
	return ae.Name + " = " + ae.Value.String()
}

// PipeExpression: value | func
type PipeExpression struct {
	Left  Expression
	Right Expression
}

func (pe *PipeExpression) expressionNode() {}
func (pe *PipeExpression) String() string {
	return pe.Left.String() + " | " + pe.Right.String()
}
