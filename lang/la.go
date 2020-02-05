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

func (c Fallthrough) HasVariable(name string) *Variable {
	return nil
}

func (c Fallthrough) Print() {
	fmt.Print("fallthrough")
}
