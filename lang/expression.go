package lang

import (
	"errors"
	"fmt"
)

type Function struct {
	parent Node

	Args []Variable
	Body Code

	Name   string
	Return string
}

func (f Function) Parent() Node {
	return f.parent
}

func (f *Function) SetParent(n Node) {
	f.parent = n
}

func (f Function) HasVariable(name string) *Variable {
	return f.DefinesVariable(name)
}

func (f *Function) DefineVariable(v Variable) {
	f.Args = append(f.Args, v)
}

func (f Function) DefinesVariable(name string) *Variable {
	for _, a := range f.Args {
		if a.Name == name {
			return &a
		}
	}
	return nil
}

func (f Function) Print() {
	fmt.Print("func " + f.Name + "(")
	for i := 0; i < len(f.Args); i++ {
		a := f.Args[i]
		fmt.Print(a.Name + " ")
		if a.Reference {
			fmt.Print("*")
		}
		fmt.Print(a.GetType())
		if i < len(f.Args)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Print(") ")

	if f.Return != Void {
		fmt.Print(f.Return + " ")
	}
	f.Body.Print()
	fmt.Print("\n")
}

func (f *Function) AddStatement(n Node) {
	f.Body.AddStatement(n)
}

func (f Function) GetType() string {
	return f.Return
}

type Const struct {
	parent Node

	Value string
}

func (c Const) Parent() Node {
	return c.parent
}

func (c *Const) SetParent(n Node) {
	c.parent = n
}

// TODO: Is this correct return type?
func (c Const) GetType() string {
	return Bool
}

func (c Const) Print() {
	fmt.Print(c.Value)
}

type Return struct {
	parent Node

	Expression Expression
}

func (r Return) Parent() Node {
	return r.parent
}

func (r *Return) SetParent(n Node) {
	r.parent = n
}

func (r Return) GetType() string {
	return r.Expression.GetType()
}

func (r Return) Print() {
	fmt.Print("return ")
	r.Expression.Print()
}

type Assign struct {
	parent Node

	left  *Variable
	right *Expression

	FirstDefinition bool
}

func (a Assign) Parent() Node {
	return a.parent
}

func (a *Assign) SetParent(n Node) {
	a.parent = n
}

func (a Assign) GetType() string {
	return a.left.Type
}

func (a Assign) Print() {
	fmt.Print(a.left.Name)
	if a.FirstDefinition {
		fmt.Print(" := ")
	} else {
		fmt.Print(" = ")
	}
	(*a.right).Print()
}

func (a Assign) Left() *Variable {
	return a.left
}

func NewAssign(left *Variable, right Expression) *Assign {
	return &Assign{
		left:  left,
		right: &right,

		FirstDefinition: false,
	}
}

type Number struct {
	parent Node

	Value string
}

func (nb Number) Parent() Node {
	return nb.parent
}

func (nb *Number) SetParent(n Node) {
	nb.parent = n
}

func (nb Number) GetType() string {
	return Int
}

func (nb Number) Print() {
	fmt.Print(nb.Value)
}

// Float and Number can be merged, only Type is different.
// There is no way how can I write down 0.0 as some nice string
// for := operator.
type Float struct {
	parent Node

	Value string
}

func (f Float) Parent() Node {
	return f.parent
}

func (f *Float) SetParent(n Node) {
	f.parent = n
}

func (f Float) GetType() string {
	return Float64
}

func (f Float) Print() {
	fmt.Print(f.Value)
}

type Str struct {
	parent Node

	Value string
}

func (s Str) Parent() Node {
	return s.parent
}

func (s *Str) SetParent(n Node) {
	s.parent = n
}

func (s Str) GetType() string {
	return String
}

func (s Str) Print() {
	fmt.Print(s.Value)
}

type Array struct {
	parent Node

	Values []Expression
	Type   string
}

func (a Array) Parent() Node {
	return a.parent
}

func (a *Array) SetParent(n Node) {
	a.parent = n
}

func (a Array) GetType() string {
	return a.Type
}

func (a Array) Print() {
	fmt.Printf("[]%s{", a.Type)
	size := len(a.Values)
	for i := 0; i < size; i++ {
		a.Values[i].Print()
		if i < size-1 {
			fmt.Print(", ")
		}

	}
	fmt.Print("}")
}

type FetchArr struct {
	parent Node

	Arr   *Variable
	Index Expression
}

func (fa FetchArr) Parent() Node {
	return fa.parent
}

func (fa *FetchArr) SetParent(n Node) {
	fa.parent = n
}

func (fa FetchArr) GetType() string {
	return fa.Arr.GetType()
}

func (fa FetchArr) Print() {
	fa.Arr.Print()
	fmt.Print("[")
	fa.Index.Print()
	fmt.Print("]")
}

type UnaryMinus struct {
	parent Node

	Expr Expression
}

func (m UnaryMinus) Parent() Node {
	return m.parent
}

func (m *UnaryMinus) SetParent(n Node) {
	m.parent = n
}

func (m UnaryMinus) GetType() string {
	return m.Expr.GetType()
}

func (m UnaryMinus) Print() {
	fmt.Print("-")
	m.Expr.Print()
}

type Negation struct {
	parent Node

	Right Expression
}

func (neg Negation) Parent() Node {
	return neg.parent
}

func (neg *Negation) SetParent(n Node) {
	neg.parent = n
}

func (neg Negation) GetType() string {
	return Bool
}

func (neg Negation) Print() {
	fmt.Print("!(")
	neg.Right.Print()
	fmt.Print(")")
}

type BinaryOp struct {
	parent Node

	inBrackets bool

	Operation string

	Right Expression
	Left  Expression
}

func (p BinaryOp) Parent() Node {
	return p.parent
}

func (p *BinaryOp) SetParent(n Node) {
	p.parent = n
}

func (p BinaryOp) GetType() string {
	if p.Right.GetType() != p.Left.GetType() {
		panic(`Left operand has different type than the right one.`)
	}

	op := p.Operation
	if op == "<" || op == "<=" || op == ">" || op == ">=" || op == "==" {
		return Bool
	}

	return p.Right.GetType()
}

// See https://golang.org/ref/spec#Operator_precedence
func (p BinaryOp) OperatorPrecedence() int {
	switch p.Operation {
	case "*":
		fallthrough
	case "/":
		fallthrough
	case "%":
		fallthrough
	case "<<":
		fallthrough
	case ">>":
		fallthrough
	case "&":
		fallthrough
	case "&^":
		return 5

	case "+":
		fallthrough
	case "-":
		fallthrough
	case "|":
		fallthrough
	case "^":
		return 4

	case "==":
		fallthrough
	case "!=":
		fallthrough
	case "<":
		fallthrough
	case "<=":
		fallthrough
	case ">":
		fallthrough
	case ">=":
		return 3

	case "&&":
		return 2

	case "||":
		return 1
	}
	panic(`Unknown type "` + p.Operation + `"`)
}

func (p BinaryOp) Print() {
	if p.inBrackets {
		fmt.Print("(")
	}
	p.Left.Print()
	fmt.Print(" " + p.Operation + " ")
	p.Right.Print()
	if p.inBrackets {
		fmt.Print(")")
	}
}

func NewBinaryOp(op string, left, right Expression) (*BinaryOp, error) {
	if left == nil {
		return nil, errors.New("Left expression is missing.")
	}
	if right == nil {
		return nil, errors.New("Left expression is missing.")
	}
	lt := left.GetType()
	rt := right.GetType()

	if lt == Void || rt == Void {
		panic(`Binary op cannot be used with "void"`)
	}

	convertToMatchingType(left, right)
	ret := &BinaryOp{
		inBrackets: false,
		Operation:  op,
		Left:       left,
		Right:      right,
	}
	left.SetParent(ret)
	right.SetParent(ret)

	bp, ok := left.(*BinaryOp)
	if ok && ret.OperatorPrecedence() > bp.OperatorPrecedence() {
		bp.inBrackets = true
	}
	bp, ok = right.(*BinaryOp)
	if ok && ret.OperatorPrecedence() > bp.OperatorPrecedence() {
		bp.inBrackets = true
	}

	return ret, nil
}

func convertToMatchingType(left, right Expression) {
	lt := left.GetType()
	rt := right.GetType()
	if lt == rt {
		return
	}

	// PHP tries to convert string to number,
	// skipping for now.
	switch lt {
	case Bool:
		switch rt {
		case Int:
			f := &FunctionCall{
				Name:   "std.BoolToInt",
				Args:   make([]Expression, 1),
				Return: Int,
			}
			f.Args[0] = left
			f.SetParent(left)
			left = f

		case Float64:
			f := &FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   make([]Expression, 1),
				Return: Float64,
			}
			f.Args[0] = left
			f.SetParent(left)
			left = f
		}

	case Int:
		switch rt {
		case Bool:
			f := &FunctionCall{
				Name:   "std.BoolToInt",
				Args:   make([]Expression, 1),
				Return: Int,
			}
			f.Args[0] = right
			f.SetParent(right)
			right = f

		case Float64:
			f := &FunctionCall{
				Name:   "float64",
				Args:   make([]Expression, 1),
				Return: Float64,
			}
			f.Args[0] = left
			f.SetParent(left)
			left = f
		}

	case Float64:
		switch rt {
		case Bool:
			f := &FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   make([]Expression, 1),
				Return: Float64,
			}
			f.Args[0] = right
			f.SetParent(right)
			right = f

		case Int:
			f := &FunctionCall{
				Name:   "float",
				Args:   make([]Expression, 1),
				Return: Float64,
			}
			f.Args[0] = right
			f.SetParent(right)
			right = f
		}
	}
}

type FunctionCall struct {
	parent Node

	Name   string
	Args   []Expression
	Return string
}

func (f *FunctionCall) AddArg(e Expression) {
	f.Args = append(f.Args, e)
}

func (f FunctionCall) Parent() Node {
	return f.parent
}

func (f *FunctionCall) SetParent(n Node) {
	f.parent = n
}

func (f FunctionCall) GetType() string {
	return f.Return
}

func (f FunctionCall) Print() {
	fmt.Print(f.Name)
	fmt.Print("(")
	for i := 0; i < len(f.Args); i++ {
		if v, isVar := f.Args[i].(*Variable); isVar {
			if v.Reference {
				fmt.Print("&")
			}
		}

		f.Args[i].Print()
		if i < len(f.Args)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Print(")")
}
