package lang

import "fmt"

type GlobalContext struct {
	parent Node

	Vars  []Variable
	Funcs []Function
}

func CreateGlobalContext() *GlobalContext {
	return &GlobalContext{
		parent: nil,
		Vars:   make([]Variable, 0),
		Funcs:  make([]Function, 0),
	}
}

func (gc *GlobalContext) Add(f Function) {
	gc.Funcs = append(gc.Funcs, f)
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

	args []Variable
	body Code

	Name   string
	Return Variable
}

func (f Function) Parent() Node {
	return f.parent
}

func (f *Function) SetParent(n Node) {
	f.parent = n
}

func (f Function) HasVariable(name string) bool {
	for _, a := range f.args {
		if a.Name == name {
			return true
		}
	}
	if f.body.HasVariable(name) {
		return true
	}

	if p := f.Parent(); p != nil {
		return p.HasVariable(name)
	}
	return false
}

// Might be good put it into the interface.
// And Code should know it too, right?
func (f *Function) DefineVariable(v Variable) {
	f.body.Vars = append(f.body.Vars, v)
}

func (f Function) Print() {
	fmt.Print("func " + f.Name + "() ")
	f.body.Print()
}

func (f *Function) AddStatement(n Node) {
	f.body.AddStatement(n)
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

func (c Code) HasVariable(name string) bool {
	for _, v := range c.Vars {
		if v.Name == name {
			return true
		}
	}
	if p := c.Parent(); p != nil {
		return p.HasVariable(name)
	}
	return false
}

func (c *Code) AddStatement(n Node) {
	n.SetParent(c)
	c.Statements = append(c.Statements, n)
}

func (c Code) Print() {
	fmt.Println("{")
	for _, s := range c.Statements {
		s.Print()
	}
	fmt.Println("}")
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

// p/p.go:69:6: cannot use lang.HTML literal (type lang.HTML) as type lang.Node in assignment:
// lang.HTML does not implement lang.Node (SetParent method has pointer receiver)
func (h /* * */ HTML) SetParent(n Node) {
	h.parent = n
}

func (h HTML) HasVariable(name string) bool {
	if h.parent != nil {
		return h.parent.HasVariable(name)
	}
	return false
}

func (h HTML) Print() {
	fmt.Println("fmt.Print(`" + h.Content + "`)")
}

func CreateMain() *Function {
	return &Function{
		Name: "main",
		body: Code{
			Vars:       make([]Variable, 0),
			Statements: make([]Node, 0),
		},
	}
}
