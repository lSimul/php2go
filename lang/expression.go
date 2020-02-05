package lang

import "fmt"

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
	for _, a := range f.Args {
		if a.Name == name {
			return &a
		}
	}

	if p := f.Parent(); p != nil {
		return p.HasVariable(name)
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

type For struct {
	parent Node

	Vars []Variable

	Init Expression
	Cond Expression
	Loop Expression

	Block *Code
}

func (f For) Parent() Node {
	return f.parent
}

func (f *For) SetParent(n Node) {
	f.parent = n
}

func (f For) HasVariable(name string) *Variable {
	v := f.DefinesVariable(name)
	if v != nil {
		return v
	}
	if f.parent != nil {
		return f.parent.HasVariable(name)
	}
	return nil
}

func (f *For) DefineVariable(v Variable) {
	f.Vars = append(f.Vars, v)
}

func (f For) DefinesVariable(name string) *Variable {
	for _, v := range f.Vars {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (f For) GetType() string {
	return Void
}

func (f *For) AddStatement(n Node) {
	f.Block.AddStatement(n)
}

func (f For) Print() {
	fmt.Print("for ")
	if f.Init != nil {
		f.Init.Print()
	}
	fmt.Print("; ")
	if f.Cond != nil {
		if f.Cond.GetType() != Bool {
			panic(`Condition does not return bool.`)
		}

		f.Cond.Print()
	}
	fmt.Print("; ")
	if f.Loop != nil {
		f.Loop.Print()
	}
	fmt.Print(" ")
	f.Block.Print()
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

func (c Const) HasVariable(name string) *Variable {
	return nil
}

// TODO: Is this correct return type?
func (c Const) GetType() string {
	return Bool
}

func (c Const) Print() {
	fmt.Print(c.Value)
}

type If struct {
	parent Node

	Vars []Variable

	Init Expression
	Cond Expression

	True  *Code
	False Expression
}

func (i If) Parent() Node {
	return i.parent
}

func (i *If) SetParent(n Node) {
	i.parent = n
}

func (i If) HasVariable(name string) *Variable {
	v := i.DefinesVariable(name)
	if v != nil {
		return v
	}
	if i.parent != nil {
		return i.parent.HasVariable(name)
	}
	return nil
}

func (i *If) DefineVariable(v Variable) {
	i.Vars = append(i.Vars, v)
}

func (i If) DefinesVariable(name string) *Variable {
	for _, v := range i.Vars {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (i If) GetType() string {
	return Void
}

func (i *If) AddStatement(n Node) {}

func (i If) Print() {
	fmt.Print("if ")
	if i.Init != nil {
		i.Init.Print()
		fmt.Print("; ")
	}
	if i.Cond != nil {
		if i.Cond.GetType() != Bool {
			panic(`Condition does not return bool.`)
		}

		i.Cond.Print()
	}
	fmt.Print(" ")
	i.True.Print()
	if i.False != nil {
		fmt.Print(" else ")
		i.False.Print()
	}
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

func (r Return) HasVariable(name string) *Variable {
	if r.parent != nil {
		return r.parent.HasVariable(name)
	}
	return nil
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

func (a Assign) HasVariable(name string) *Variable {
	if a.parent != nil {
		return a.parent.HasVariable(name)
	}
	return nil
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

func CreateAssign(left *Variable, right Expression) *Assign {
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

func (nb Number) HasVariable(name string) *Variable {
	return nil
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

func (f Float) HasVariable(name string) *Variable {
	return nil
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

func (s Str) HasVariable(name string) *Variable {
	return nil
}

func (s Str) GetType() string {
	return String
}

func (s Str) Print() {
	fmt.Print(s.Value)
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

func (m UnaryMinus) HasVariable(name string) *Variable {
	if m.parent != nil {
		return m.parent.HasVariable(name)
	}
	return nil
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

func (neg Negation) HasVariable(name string) *Variable {
	if neg.parent != nil {
		return neg.parent.HasVariable(name)
	}
	return nil
}

func (neg Negation) GetType() string {
	return neg.Right.GetType()
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

func (p BinaryOp) HasVariable(name string) *Variable {
	if p.parent != nil {
		return p.parent.HasVariable(name)
	}
	return nil
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

func CreateBinaryOp(op string, left, right Expression) *BinaryOp {
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

	return ret
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

type Code struct {
	parent Node

	Vars       []Variable
	Statements []Node
}

func (c Code) Parent() Node {
	return c.parent
}

func (c *Code) SetParent(n Node) {
	c.parent = n
}

func (c Code) HasVariable(name string) *Variable {
	v := c.DefinesVariable(name)
	if v != nil {
		return v
	}
	if p := c.Parent(); p != nil {
		return p.HasVariable(name)
	}
	return nil
}

func (c Code) GetType() string {
	return Void
}

func (c *Code) DefineVariable(v Variable) {
	c.Vars = append(c.Vars, v)
}

func (c Code) DefinesVariable(name string) *Variable {
	for _, v := range c.Vars {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (c *Code) AddStatement(n Node) {
	n.SetParent(c)
	c.Statements = append(c.Statements, n)
}

func (c Code) Print() {
	fmt.Print("{\n")
	for _, s := range c.Statements {
		s.Print()
		fmt.Print("\n")
	}
	fmt.Print("}")
}

type Switch struct {
	parent Node

	Condition Expression
	// Default will end up here too,
	// to keep the order from PHP.
	Cases []Node
}

func (sw Switch) Parent() Node {
	return sw.parent
}

func (sw *Switch) SetParent(n Node) {
	sw.parent = n
}

func (sw Switch) HasVariable(name string) *Variable {
	if sw.parent != nil {
		return sw.parent.HasVariable(name)
	}
	return nil
}

func (sw Switch) GetType() string {
	return sw.Condition.GetType()
}

func (sw *Switch) AddStatement(n Node)       {}
func (sw *Switch) DefineVariable(v Variable) {}

func (sw Switch) DefinesVariable(name string) *Variable {
	return nil
}

func (sw Switch) Print() {
	fmt.Print("switch ")
	sw.Condition.Print()
	fmt.Print(" {\n")
	for _, c := range sw.Cases {
		c.Print()
	}
	fmt.Print("}")
}

type Case struct {
	parent *Switch

	Statements []Node
	Vars       []Variable
	Condition  Expression
}

func (c Case) Parent() Node {
	return c.parent
}

func (c *Case) SetParent(n Node) {
	sw, ok := n.(*Switch)
	if !ok {
		panic(`Expected pointer to switch, something else given.`)
	}
	c.parent = sw
}

func (c Case) HasVariable(name string) *Variable {
	v := c.DefinesVariable(name)
	if v != nil {
		return v
	}
	if c.parent != nil {
		return c.parent.HasVariable(name)
	}
	return nil
}

func (c Case) GetType() string {
	return c.Condition.GetType()
}

func (c Case) Print() {
	fmt.Print("case ")
	c.Condition.Print()
	fmt.Print(":\n")
	for _, e := range c.Statements {
		e.Print()
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

func (c *Case) AddStatement(n Node) {
	c.Statements = append(c.Statements, n)
}

func (c *Case) DefineVariable(v Variable) {
	c.Vars = append(c.Vars, v)
}

func (c Case) DefinesVariable(name string) *Variable {
	for _, v := range c.Vars {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

type Default struct {
	parent *Switch

	Vars       []Variable
	Statements []Node
}

func (d Default) Parent() Node {
	return d.parent
}

func (d *Default) SetParent(n Node) {
	sw, ok := n.(*Switch)
	if !ok {
		panic(`Expected pointer to switch, something else given.`)
	}
	d.parent = sw
}

func (d Default) HasVariable(name string) *Variable {
	v := d.DefinesVariable(name)
	if v != nil {
		return v
	}
	if d.parent != nil {
		return d.parent.HasVariable(name)
	}
	return nil
}

func (d Default) GetType() string {
	return Void
}

func (d Default) Print() {
	fmt.Print("default:\n")
	for _, e := range d.Statements {
		e.Print()
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

func (d *Default) AddStatement(n Node) {
	d.Statements = append(d.Statements, n)
}

func (d *Default) DefineVariable(v Variable) {
	d.Vars = append(d.Vars, v)
}

func (d Default) DefinesVariable(name string) *Variable {
	for _, v := range d.Vars {
		if v.Name == name {
			return &v
		}
	}
	return nil
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

func (f FunctionCall) HasVariable(name string) *Variable {
	if f.parent != nil {
		return f.parent.HasVariable(name)
	}
	return nil
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

type Inc struct {
	parent Node

	Var *Variable
}

func (i Inc) Parent() Node {
	return i.parent
}

func (i *Inc) SetParent(n Node) {
	i.parent = n
}

func (i Inc) HasVariable(name string) *Variable {
	if i.parent != nil {
		return i.parent.HasVariable(name)
	}
	return nil
}

func (i Inc) GetType() string {
	return i.Var.GetType()
}

func (i Inc) Print() {
	if i.Var.Reference {
		fmt.Print("(*")
		i.Var.Print()
		fmt.Print(")")
	} else {
		i.Var.Print()
	}
	fmt.Print("++")
}

type Dec struct {
	parent Node

	Var *Variable
}

func (d Dec) Parent() Node {
	return d.parent
}

func (d *Dec) SetParent(n Node) {
	d.parent = n
}

func (d Dec) HasVariable(name string) *Variable {
	if d.parent != nil {
		return d.parent.HasVariable(name)
	}
	return nil
}

func (d Dec) GetType() string {
	return d.Var.GetType()
}

func (d Dec) Print() {
	d.Var.Print()
	fmt.Print("--")
}
