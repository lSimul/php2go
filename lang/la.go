package lang

import "fmt"

type GlobalContext struct {
	parent Node

	Vars  []Variable
	Funcs map[string]*Function
}

func CreateGlobalContext() *GlobalContext {
	return &GlobalContext{
		parent: nil,
		Vars:   make([]Variable, 0),
		Funcs:  make(map[string]*Function, 0),
	}
}

func (gc *GlobalContext) Add(f *Function) {
	gc.Funcs[f.Name] = f
}

func (gc GlobalContext) Get(name string) *Function {
	return gc.Funcs[name]
}

func (gc GlobalContext) Print() {
	fmt.Println("package main")
	fmt.Println()
	fmt.Println("import \"fmt\"")
	fmt.Println()

	for _, f := range gc.Funcs {
		f.Print()
	}
}

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
		fmt.Print(a.Name + " " + a.GetType())
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
	for _, v := range c.Vars {
		if v.Name == name {
			return &v
		}
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

// Refactor to something like function call
// function call fmt.Println or something
// different what should be in the standard
// library.
type HTML struct {
	parent  Node
	Content string
}

func (h HTML) Parent() Node {
	return h.parent
}

func (h *HTML) SetParent(n Node) {
	h.parent = n
}

func (h HTML) HasVariable(name string) *Variable {
	if h.parent != nil {
		return h.parent.HasVariable(name)
	}
	return nil
}

func (h HTML) Print() {
	fmt.Print("fmt.Print(`" + h.Content + "`)")
}

func CreateFunc(name string) *Function {
	f := &Function{
		Name: name,
		Body: Code{
			Vars:       make([]Variable, 0),
			Statements: make([]Node, 0),
		},
		Return: Void,
	}
	f.Body.SetParent(f)
	return f
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
	for _, v := range f.Vars {
		if v.Name == name {
			return &v
		}
	}
	if f.parent != nil {
		return f.parent.HasVariable(name)
	}
	return nil
}

func (f *For) DefineVariable(v Variable) {
	f.Vars = append(f.Vars, v)
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
	for _, v := range i.Vars {
		if v.Name == name {
			return &v
		}
	}
	if i.parent != nil {
		return i.parent.HasVariable(name)
	}
	return nil
}

func (i *If) DefineVariable(v Variable) {
	i.Vars = append(i.Vars, v)
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

func (a Number) Parent() Node {
	return a.parent
}

func (a *Number) SetParent(n Node) {
	a.parent = n
}

func (n Number) HasVariable(name string) *Variable {
	return nil
}

func (n Number) GetType() string {
	return Int
}

func (n Number) Print() {
	fmt.Print(n.Value)
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

func (a Str) Parent() Node {
	return a.parent
}

func (a *Str) SetParent(n Node) {
	a.parent = n
}

func (n Str) HasVariable(name string) *Variable {
	return nil
}

func (n Str) GetType() string {
	return String
}

func (n Str) Print() {
	fmt.Print(n.Value)
}

type UnaryMinus struct {
	parent Node

	Right Expression
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
	return m.Right.GetType()
}

func (m UnaryMinus) Print() {
	fmt.Print("-")
	m.Right.Print()
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

func (p BinaryOp) Print() {
	p.Left.Print()
	fmt.Print(" " + p.Operation + " ")
	p.Right.Print()
}

func CreateBinaryOp(op string, left, right Expression) *BinaryOp {
	return &BinaryOp{
		Operation: op,
		Left:      left,
		Right:     right,
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
	i.Var.Print()
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

type Break struct {
	parent Node
}

func (b Break) Parent() Node {
	return b.parent
}

func (b *Break) SetParent(n Node) {
	b.parent = n
}

func (b Break) HasVariable(name string) *Variable {
	return nil
}

func (b Break) GetType() string {
	return Void
}

func (b Break) Print() {
	fmt.Print("break")
}

type Continue struct {
	parent Node
}

func (c Continue) Parent() Node {
	return c.parent
}

func (c *Continue) SetParent(n Node) {
	c.parent = n
}

func (c Continue) HasVariable(name string) *Variable {
	return nil
}

func (c Continue) GetType() string {
	return Void
}

func (c Continue) Print() {
	fmt.Print("continue")
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

func (c Const) GetType() string {
	return Bool
}

func (c Const) Print() {
	fmt.Print(c.Value)
}
