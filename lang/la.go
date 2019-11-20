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
	body Code

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
	if v := f.body.HasVariable(name); v != nil {
		return v
	}

	if p := f.Parent(); p != nil {
		return p.HasVariable(name)
	}
	return nil
}

// Might be good put it into the interface.
// And Code should know it too, right?
func (f *Function) DefineVariable(v Variable) {
	f.body.Vars = append(f.body.Vars, v)
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
	f.body.Print()
	fmt.Print("\n")
}

func (f *Function) AddStatement(n Node) {
	f.body.AddStatement(n)
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
	fmt.Print("}\n")
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
	return &Function{
		Name: name,
		body: Code{
			Vars:       make([]Variable, 0),
			Statements: make([]Node, 0),
		},
		Return: Void,
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

	Value int
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
	return p.Right.GetType()
}

func (p BinaryOp) Print() {
	p.Left.Print()
	fmt.Print(" " + p.Operation + " ")
	p.Right.Print()
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

// TODO: This needs to be solved
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
