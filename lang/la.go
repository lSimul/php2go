package lang

import (
	"errors"
	"fmt"
	"strings"
)

type GlobalContext struct {
	parent Node

	vars  []*Variable
	Funcs map[string]*Function

	namespaces map[string]string
}

func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		parent: nil,
		vars:   make([]*Variable, 0),
		Funcs:  make(map[string]*Function, 0),

		namespaces: make(map[string]string),
	}
}

// TODO: Pull this out of this package, it is starting to be
// very complicated.
func (gc *GlobalContext) RequireNamespace(n string) {
	if _, ok := gc.namespaces[n]; !ok {
		switch n {
		case "fmt":
			gc.namespaces[n] = "fmt"
		case "std":
			gc.namespaces[n] = "php2go/std"
		case "array":
			gc.namespaces[n] = "php2go/std/array"

		default:
			panic(`Unknown namespace`)
		}
	}
}

func (gc GlobalContext) SetParent(n Node) {}
func (gc GlobalContext) Parent() Node     { return nil }

func (gc GlobalContext) AddStatement(n Node) { panic(`not implemented`) }

func (gc *GlobalContext) DefineVariable(v *Variable) {
	for _, vr := range gc.vars {
		if vr.Name == v.Name {
			vr.typ = Anything
			return
		}
	}
	gc.vars = append(gc.vars, v)
}

func (gc GlobalContext) HasVariable(name string) *Variable {
	return gc.DefinesVariable(name)
}

func (gc GlobalContext) DefinesVariable(name string) *Variable {
	for _, v := range gc.vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (gc *GlobalContext) Add(f *Function) {
	gc.Funcs[f.Name] = f
}

func (gc GlobalContext) Get(name string) *Function {
	return gc.Funcs[name]
}

func (gc GlobalContext) String() string {
	s := strings.Builder{}
	s.WriteString("package main\n\n")

	if len(gc.namespaces) == 1 {
		for _, n := range gc.namespaces {
			s.WriteString("import \"" + n + "\"\n")
		}
	} else if len(gc.namespaces) > 0 {
		s.WriteString("import (\n")
		for _, n := range gc.namespaces {
			s.WriteString("\"" + n + "\"\n")
		}
		s.WriteString(")\n\n")
	}

	for _, v := range gc.vars {
		s.WriteString(fmt.Sprintf("var %s %s\n", v, v.Type()))
	}
	if len(gc.vars) > 0 {
		s.WriteByte('\n')
	}

	for _, f := range gc.Funcs {
		s.WriteString(f.String())
	}
	return s.String()
}

func NewFunc(name string) *Function {
	f := &Function{
		Name: name,
		Body: Code{
			Vars:       make([]*Variable, 0),
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

func (b Break) String() string {
	return "break"
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

func (c Continue) String() string {
	return "continue"
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

func (c Fallthrough) String() string {
	return "fallthrough"
}

type Code struct {
	parent Block

	Vars       []*Variable
	Statements []Node
}

func (c Code) Parent() Node {
	return c.parent
}

func (c *Code) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	c.parent = n.(Block)
}

func (c Code) HasVariable(name string) *Variable {
	v := c.DefinesVariable(name)
	if v != nil {
		return v
	}
	if p := c.parent; p != nil {
		return p.HasVariable(name)
	}
	return nil
}

func (c *Code) DefineVariable(v *Variable) {
	// TODO: Definition should be small, this does a lot of things.
	for _, vr := range c.Vars {
		if vr != v {
			continue
		}
		if vr.typ == Anything {
			return
		}
		vr.typ = Anything

		switch fd := vr.FirstDefinition.(type) {
		case *VarDef:

		case *Assign:
			fd.FirstDefinition = false
			vd := newVarDef(c, vr)
			vr.FirstDefinition = vd
			c.Statements = append([]Node{vd}, c.Statements...)
			return

		default:
			panic(`Wrong assignment.`)
		}
	}
	c.Vars = append(c.Vars, v)
}

func (c Code) DefinesVariable(name string) *Variable {
	for _, v := range c.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (c *Code) AddStatement(n Node) {
	n.SetParent(c)
	c.Statements = append(c.Statements, n)
}

func (c Code) String() string {
	s := strings.Builder{}
	s.WriteString("{\n")
	for _, st := range c.Statements {
		s.WriteString(st.String())
		s.WriteString("\n")
	}
	s.WriteString("}")
	return s.String()
}

func NewCode(parent Node) *Code {
	// TODO: Be loud, return an error
	// instead of staying silent.
	var p Block
	if b, ok := parent.(Block); ok {
		p = b
	}
	return &Code{
		parent:     p,
		Vars:       make([]*Variable, 0),
		Statements: make([]Node, 0),
	}
}

type For struct {
	parent Block

	Vars []*Variable

	Init Node
	cond Expression
	Loop Node

	Block *Code
}

func (f *For) SetCond(e Expression) error {
	if e.Type() != Bool {
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

func (f *For) DefineVariable(v *Variable) {
	for _, vr := range f.Vars {
		if vr.Name != v.Name || vr.typ == v.typ {
			continue
		}
		vr.typ = Anything
		a, ok := vr.FirstDefinition.(*Assign)
		if !ok {
			panic(`For cycle cannot move to VarDef.`)
		}
		a.FirstDefinition = false
		return
	}
	f.Vars = append(f.Vars, v)
}

func (f For) DefinesVariable(name string) *Variable {
	for _, v := range f.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (f *For) AddStatement(n Node) {
	f.Block.AddStatement(n)
}

func (f For) String() string {
	s := strings.Builder{}
	s.WriteString("for ")
	if f.Init != nil {
		s.WriteString(f.Init.String())
	}
	s.WriteString("; ")
	if f.cond != nil {
		s.WriteString(f.cond.String())
	}
	s.WriteString("; ")
	if f.Loop != nil {
		s.WriteString(f.Loop.String())
	}
	s.WriteString(" ")
	s.WriteString(f.Block.String())
	return s.String()
}

func ConstructFor(parent Block) *For {
	f := &For{
		parent: parent,
		Vars:   make([]*Variable, 0),
		Block: &Code{
			Vars:       make([]*Variable, 0),
			Statements: make([]Node, 0),
		},
	}
	f.Block.SetParent(f)
	return f
}

// TODO: This does not work as it should,
// that variable assignment is done just by
// a simple string, I do not care if these
// values are used later on. This makes sense
// if I want to use this structure for other
// languages, too. Range is fairly specific.
type Foreach struct {
	parent Block

	Iterated Expression

	Key   *Variable
	Value Variable

	Block *Code
}

func (f Foreach) Parent() Node {
	return f.parent
}

func (f *Foreach) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	f.parent = n.(Block)
}

func (f Foreach) HasVariable(name string) *Variable {
	v := f.DefinesVariable(name)
	if v != nil {
		return v
	}

	if f.parent != nil {
		return f.parent.HasVariable(name)
	}
	return nil
}

func (f *Foreach) AddStatement(n Node)        {}
func (f *Foreach) DefineVariable(v *Variable) {}

func (f Foreach) DefinesVariable(name string) *Variable {
	if f.Key != nil && f.Key.Name == name {
		return f.Key
	}
	if f.Value.Name == name {
		return &f.Value
	}
	return nil
}

func (f Foreach) String() string {
	k := "_"
	if f.Key != nil {
		k = f.Key.Name
	}
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("for %s, %s := range ", k, f.Value.Name))
	s.WriteString(f.Iterated.String())
	s.WriteString(" ")
	s.WriteString(f.Block.String())
	return s.String()
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

func (sw *Switch) AddStatement(n Node)        {}
func (sw *Switch) DefineVariable(v *Variable) {}

func (sw Switch) DefinesVariable(name string) *Variable {
	return nil
}

func (sw Switch) String() string {
	s := strings.Builder{}
	s.WriteString("switch ")
	s.WriteString(sw.Condition.String())
	s.WriteString(" {\n")
	for _, c := range sw.Cases {
		s.WriteString(c.String())
	}
	s.WriteByte('}')
	return s.String()
}

type Case struct {
	parent *Switch

	Statements []Node
	Vars       []*Variable
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

func (c Case) String() string {
	s := strings.Builder{}
	s.WriteString("case ")
	s.WriteString(c.Condition.String())
	s.WriteString(":\n")
	for _, e := range c.Statements {
		s.WriteString(e.String())
		s.WriteByte('\n')
	}
	s.WriteByte('\n')
	return s.String()
}

func (c *Case) AddStatement(n Node) {
	c.Statements = append(c.Statements, n)
}

func (c *Case) DefineVariable(v *Variable) {
	// TODO: Definition should be small, this does a lot of things.
	for _, vr := range c.Vars {
		if vr != v {
			continue
		}
		if vr.Type() == Anything {
			return
		}
		vr.typ = Anything

		switch vr.FirstDefinition.(type) {
		case *VarDef:

		case *Assign:
			vr.FirstDefinition.(*Assign).FirstDefinition = false
			vd := newVarDef(c, vr)
			vr.FirstDefinition = vd
			c.Statements = append([]Node{vd}, c.Statements...)
			return

		default:
			panic(`Wrong assignment.`)
		}
	}
	c.Vars = append(c.Vars, v)
}

func (c Case) DefinesVariable(name string) *Variable {
	for _, v := range c.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type Default struct {
	parent *Switch

	Vars       []*Variable
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

func (d Default) String() string {
	s := strings.Builder{}
	s.WriteString("default:\n")
	for _, e := range d.Statements {
		s.WriteString(e.String())
		s.WriteByte('\n')
	}
	s.WriteByte('\n')
	return s.String()
}

func (d *Default) AddStatement(n Node) {
	d.Statements = append(d.Statements, n)
}

func (d *Default) DefineVariable(v *Variable) {
	// TODO: Definition should be small, this does a lot of things.
	for _, vr := range d.Vars {
		if vr != v {
			continue
		}
		if vr.Type() == Anything {
			return
		}
		vr.typ = Anything

		switch vr.FirstDefinition.(type) {
		case *VarDef:

		case *Assign:
			vr.FirstDefinition.(*Assign).FirstDefinition = false
			vd := newVarDef(d, vr)
			vr.FirstDefinition = vd
			d.Statements = append([]Node{vd}, d.Statements...)
			return

		default:
			panic(`Wrong assignment.`)
		}
	}
	d.Vars = append(d.Vars, v)
}

func (d Default) DefinesVariable(name string) *Variable {
	for _, v := range d.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

type If struct {
	parent Block

	Vars []*Variable

	Init Expression
	cond Expression

	True  *Code
	False Block
}

func (i *If) SetCond(e Expression) error {
	if e.Type() != Bool {
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

func (i *If) DefineVariable(v *Variable) {
	for _, vr := range i.Vars {
		if vr.Name != v.Name || vr.Type() == v.typ {
			continue
		}
		vr.typ = Anything
		a, ok := vr.FirstDefinition.(*Assign)
		if !ok {
			panic(`For cycle cannot move to VarDef.`)
		}
		a.FirstDefinition = false
		return
	}
	i.Vars = append(i.Vars, v)
}

func (i If) DefinesVariable(name string) *Variable {
	for _, v := range i.Vars {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (i *If) AddStatement(n Node) {}

func (i If) String() string {
	s := strings.Builder{}
	s.WriteString("if ")
	if i.Init != nil {
		s.WriteString(i.Init.String())
		s.WriteString("; ")
	}
	if i.cond != nil {
		s.WriteString(i.cond.String())
	}
	s.WriteByte(' ')
	s.WriteString(i.True.String())
	if i.False != nil {
		s.WriteString(" else ")
		s.WriteString(i.False.String())
	}
	return s.String()
}

type Inc struct {
	parent Node

	v *VarRef
}

func (i Inc) Parent() Node {
	return i.parent
}

func (i *Inc) SetParent(n Node) {
	i.parent = n
}

func (i Inc) String() string {
	s := strings.Builder{}
	if i.v.V.Type() == Anything {
		if i.v.typ == String {
			panic(`Unable to use Inc with 'string'.`)
		}
		s.WriteString(i.v.V.String())
		s.WriteString(" = ")
		s.WriteString(i.v.String())
		s.WriteString(" + 1")
	} else {
		if i.v.Reference {
			s.WriteByte('*')
		}
		s.WriteString(i.v.String())
		s.WriteString("++")
	}
	return s.String()
}

func NewInc(parent Node, v *VarRef) *Inc {
	return &Inc{
		parent: parent,
		v:      v,
	}
}

type Dec struct {
	parent Node

	v *VarRef
}

func (d Dec) Parent() Node {
	return d.parent
}

func (d *Dec) SetParent(n Node) {
	d.parent = n
}

func (i Dec) String() string {
	s := strings.Builder{}
	if i.v.V.Type() == Anything {
		if i.v.typ == String {
			panic(`Unable to use Dec with 'string'.`)
		}
		s.WriteString(i.v.V.String())
		s.WriteString(" = ")
		s.WriteString(i.v.String())
		s.WriteString(" - 1")
	} else {
		if i.v.Reference {
			s.WriteByte('*')
		}
		s.WriteString(i.v.String())
		s.WriteString("--")
	}
	return s.String()
}

func NewDec(parent Node, v *VarRef) *Dec {
	return &Dec{
		parent: parent,
		v:      v,
	}
}
