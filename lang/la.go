package lang

import (
	"errors"
	"fmt"
)

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
	fmt.Println("import \"php2go/std\"")
	fmt.Println()

	for _, f := range gc.Funcs {
		f.Print()
	}
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

type Break struct {
	parent Node
}

func (b Break) Parent() Node {
	return b.parent
}

func (b *Break) SetParent(n Node) {
	b.parent = n
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

func (c Continue) Print() {
	fmt.Print("continue")
}

type Fallthrough struct {
	parent Node
}

func (c Fallthrough) Parent() Node {
	return c.parent
}

func (c *Fallthrough) SetParent(n Node) {
	c.parent = n
}

func (c Fallthrough) Print() {
	fmt.Print("fallthrough")
}

type For struct {
	parent Block

	Vars []Variable

	Init Node
	cond Expression
	Loop Node

	Block *Code
}

func (f *For) SetCond(e Expression) error {
	if e.GetType() != Bool {
		return errors.New(`Condition must be an expression returning bool.`)
	}
	f.cond = e
	return nil
}

func (f For) Parent() Node {
	return f.parent
}

func (f *For) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	f.parent = n.(Block)
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

func (f *For) AddStatement(n Node) {
	f.Block.AddStatement(n)
}

func (f For) Print() {
	fmt.Print("for ")
	if f.Init != nil {
		f.Init.Print()
	}
	fmt.Print("; ")
	if f.cond != nil {
		f.cond.Print()
	}
	fmt.Print("; ")
	if f.Loop != nil {
		f.Loop.Print()
	}
	fmt.Print(" ")
	f.Block.Print()
}

func ConstructFor(parent Block) *For {
	f := &For{
		parent: parent,
		Vars:   make([]Variable, 0),
		Block: &Code{
			Vars:       make([]Variable, 0),
			Statements: make([]Node, 0),
		},
	}
	f.Block.SetParent(f)
	return f
}

type Switch struct {
	parent Block

	Condition Expression
	// Default will end up here too,
	// to keep the order from PHP.
	Cases []Node
}

func (sw Switch) Parent() Node {
	return sw.parent
}

func (sw *Switch) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	sw.parent = n.(Block)
}

func (sw Switch) HasVariable(name string) *Variable {
	if sw.parent != nil {
		return sw.parent.HasVariable(name)
	}
	return nil
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

type If struct {
	parent Block

	Vars []Variable

	Init Expression
	cond Expression

	True  *Code
	False Block
}

func (i *If) SetCond(e Expression) error {
	if e.GetType() != Bool {
		return errors.New(`Condition must be an expression returning bool.`)
	}
	i.cond = e
	return nil
}

func (i If) Parent() Node {
	return i.parent
}

func (i *If) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	i.parent = n.(Block)
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

func (i *If) AddStatement(n Node) {}

func (i If) Print() {
	fmt.Print("if ")
	if i.Init != nil {
		i.Init.Print()
		fmt.Print("; ")
	}
	if i.cond != nil {
		i.cond.Print()
	}
	fmt.Print(" ")
	i.True.Print()
	if i.False != nil {
		fmt.Print(" else ")
		i.False.Print()
	}
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

func (d Dec) Print() {
	d.Var.Print()
	fmt.Print("--")
}
