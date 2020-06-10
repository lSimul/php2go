package lang

import (
	"errors"
	"fmt"
	"strings"
)

type GlobalContext struct {
	Path string

	Files []*File

	Vars []*Variable
}

func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		Files: make([]*File, 0),
		Vars:  make([]*Variable, 0),
	}
}

func (gc *GlobalContext) Add(f *File) {
	gc.Files = append(gc.Files, f)
}

func (gc *GlobalContext) DefineVariable(v *Variable) {
	for _, vr := range gc.Vars {
		if strings.TrimPrefix(vr.Name, "g.") == strings.TrimPrefix(v.Name, "g.") {
			vr.typ = NewTyp(Anything, false)
			return
		}
	}
	v.Name = "g." + v.Name
	gc.Vars = append(gc.Vars, v)
}

func (gc *GlobalContext) HasVariable(name string, oos bool) *Variable {
	for _, v := range gc.Vars {
		if strings.TrimPrefix(v.Name, "g.") == strings.TrimPrefix(name, "g.") {
			return v
		}
	}
	return nil
}

func (gc GlobalContext) String() string {
	return gc.Files[0].String()
}

type File struct {
	parent *GlobalContext

	Name string

	vars    []*Variable
	vardefs []*VarDef
	Funcs   map[string]*Function
	imports []string

	server   bool
	withMain bool

	Main *Function
}

func NewFile(gc *GlobalContext, name string, server, withMain bool) *File {
	f := &File{
		parent: gc,

		Name: name,

		vars:    make([]*Variable, 0),
		vardefs: make([]*VarDef, 0),

		Funcs:   make(map[string]*Function, 0),
		imports: make([]string, 0),

		server:   server,
		withMain: withMain,
	}
	gc.Add(f)
	return f
}

func (f File) Parent() Node     { return nil }
func (f File) SetParent(n Node) {}

func (f *File) DefineVariable(v *Variable) {
	for _, vr := range f.vars {
		if vr.Name == v.Name {
			vr.typ = NewTyp(Anything, false)
			return
		}
	}
	vd := &VarDef{
		parent: f,
		V:      v,

		typ: v.typ.typ,
	}
	v.FirstDefinition = vd
	f.vars = append(f.vars, v)
	f.vardefs = append(f.vardefs, vd)
}

func (f File) HasVariable(name string, oos bool) *Variable {
	for _, v := range f.vars {
		if v.Name == name {
			return v
		}
	}
	if f.parent != nil {
		return f.parent.HasVariable(name, oos)
	}
	return nil
}

func (f *File) Add(fc *Function) {
	f.Funcs[fc.Name] = fc
}

func (f *File) AddImport(name string) {
	if name == "" {
		return
	}
	for _, i := range f.imports {
		if i == name {
			return
		}
	}
	f.imports = append(f.imports, name)
}

func (f *File) String() string {
	s := strings.Builder{}

	fn := strings.Builder{}
	for _, f := range f.Funcs {
		fn.WriteString(f.String())
	}

	s.WriteString("package main\n\n")

	if len(f.imports) > 0 {
		s.WriteString("import (\n")
		for _, n := range f.imports {
			s.WriteString("\"" + n + "\"\n")
		}
		s.WriteString(")\n\n")
	}

	for _, v := range f.vardefs {
		s.WriteString(v.String() + "\n")
	}

	if f.withMain {
		s.WriteString("\ntype global struct {\n")
		for _, v := range f.parent.Vars {
			s.WriteString(fmt.Sprintf("\t%s %s\n", strings.TrimPrefix(v.Name, "g."), v.typ))
		}
		s.WriteString("}\n")

		gc := f.parent
		if f.server {
			s.WriteString(`
func main() {
	flag.Parse()

	if *server != "" {
		mux := http.NewServeMux()
`)
			simpleIndex := true
			main := ""
			for _, fl := range gc.Files {
				p := strings.TrimPrefix(fl.Name, gc.Path)
				if p == "index.php" {
					main = fl.Main.Name
					simpleIndex = false
					continue
				}
				s.WriteString(`
		mux.HandleFunc("/` + p + `", func(w http.ResponseWriter, r *http.Request) {
			g := &global{
				_GET: array.NewString(),
				W: w,
			}
			for k, v := range r.URL.Query() {
				g._GET.Edit(array.NewScalar(k), v[len(v)-1])
			}
			g.` + fl.Main.Name + `()
		})`)
				if strings.HasSuffix(p, "index.php") {
					p = strings.TrimSuffix(p, "index.php")
					s.WriteString(`
					mux.HandleFunc("/` + p + `", func(w http.ResponseWriter, r *http.Request) {
						g := &global{
							_GET: array.NewString(),
							W: w,
						}
						for k, v := range r.URL.Query() {
							g._GET.Edit(array.NewScalar(k), v[len(v)-1])
						}
						g.` + fl.Main.Name + `()
					})`)
				}
			}

			if !simpleIndex {
				s.WriteString(`
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					g := &global{
						_GET: array.NewString(),
						W: w,
					}
					if r.URL.Path == "/" || r.URL.Path == "/index.php" {
						for k, v := range r.URL.Query() {
							g._GET.Edit(array.NewScalar(k), v[len(v)-1])
						}

						g.` + main + `()
						return
					}
					http.FileServer(http.Dir(".")).ServeHTTP(w, r)
				})`)

			} else {
				s.WriteString(`
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.FileServer(http.Dir(".")).ServeHTTP(w, r)
				})`)
			}

			s.WriteString(`
		// Validate that server address
		log.Fatal(http.ListenAndServe(*server, mux))
	} else {
		mainCLI()
	}
}

func mainCLI() {
	g := &global{
		_GET: array.NewString(),
		W: os.Stdout,
	}
	switch *file {
`)
			for _, fl := range gc.Files {
				p := strings.TrimPrefix(fl.Name, gc.Path)
				s.WriteString(`
	case "` + p + `":
		g.` + fl.Main.Name + `()
`)
			}
			s.WriteString(`
	default:
		g.` + f.Main.Name + `()
	}
}
`)
		} else {
			s.WriteString(`
func main() {
	g := &global{}
	switch *file {
`)
			for _, fl := range gc.Files {
				p := strings.TrimPrefix(fl.Name, gc.Path)
				s.WriteString(`
	case "` + p + `":
		g.` + fl.Main.Name + `()
`)
			}
			s.WriteString(`
	default:
		g.` + f.Main.Name + `()
	}
}
`)
		}
	}

	s.WriteString(fn.String())

	return s.String()
}

func NewFunc(name string) *Function {
	f := &Function{
		Name:          name,
		VariadicCount: false,
		Body: Code{
			Vars:       make([]*Variable, 0),
			Statements: make([]Node, 0),

			withBrackets: true,
		},

		NeedsGlobal: false,

		Return: NewTyp(Void, false),
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

	withBrackets bool
}

func (c Code) Parent() Node {
	return c.parent
}

func (c *Code) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	c.parent = n.(Block)
}

func (c *Code) HasVariable(name string, oos bool) *Variable {
	if oos {
		for i := len(c.Statements) - 1; i >= 0; i-- {
			switch s := c.Statements[i].(type) {
			case Block:
				v := s.definesVariable(name)
				if v == nil {
					continue
				}
				for _, cv := range c.Vars {
					if cv.Name == name {
						// This is a nasty hack. Instead of changing
						// VarDefs pointing to "v" I change its type.
						// Fixing this will not be that hard, during every
						// NewVarDef I will append to variable array.
						v.typ = NewTyp(Anything, false)
						cv.CurrentType = v.CurrentType
						c.DefineVariable(cv)
						return cv
					}
				}
				c.DefineVariable(v)
				vd := newVarDef(c, v)
				v.FirstDefinition = vd
				c.Statements = append([]Node{vd}, c.Statements...)
				c.Vars = append(c.Vars, v)
				return v

			case *Assign:
				if strings.TrimPrefix(s.left.Name, "g.") ==
					strings.TrimPrefix(name, "g.") {
					return c.HasVariable(name, false)
				}
			}
		}
	}

	for _, v := range c.Vars {
		if v.Name == name {
			return v
		}
	}

	if p := c.parent; p != nil {
		if v := p.HasVariable(name, oos); v != nil {
			return v
		}
	}
	return nil
}

func (c *Code) DefineVariable(v *Variable) {
	// TODO: Definition should be small, this does a lot of things.
	for _, vr := range c.Vars {
		if vr != v {
			continue
		}
		if vr.typ.Equal(Anything) {
			return
		}
		vr.typ = NewTyp(Anything, false)

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

func (c *Code) definesVariable(name string) *Variable {
	for i, v := range c.Vars {
		if v.Name == name {
			c.unset(i)
			return v
		}
	}

	for i := len(c.Statements) - 1; i >= 0; i-- {
		if b, ok := c.Statements[i].(Block); ok {
			if v := b.definesVariable(name); v != nil {
				return v
			}
		}
	}
	return nil
}

func (c *Code) unset(index int) {
	v := c.Vars[index]
	copy(c.Vars[index:], c.Vars[index+1:])
	c.Vars = c.Vars[:len(c.Vars)-1]
	for i, s := range c.Statements {
		if v.FirstDefinition != s {
			continue
		}
		switch a := v.FirstDefinition.(type) {
		case *Assign:
			a.FirstDefinition = false
		case *VarDef:
			copy(c.Statements[i:], c.Statements[i+1:])
			c.Statements = c.Statements[:len(c.Statements)-1]
		}
	}
}

func (c *Code) AddStatement(n Node) {
	n.SetParent(c)
	c.Statements = append(c.Statements, n)
}

func (c Code) String() string {
	s := strings.Builder{}
	if c.withBrackets {
		s.WriteString("{\n")
	}
	for _, st := range c.Statements {
		s.WriteString(st.String())
		s.WriteString("\n")
	}
	if c.withBrackets {
		s.WriteString("}")
	}
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

		withBrackets: true,
	}
}

type For struct {
	parent Block

	Vars []*Variable

	Init Node
	cond Expression
	Loop Node

	Block *Code

	Labels []Const
}

func (f *For) SetCond(e Expression) error {
	if !e.Type().Equal(Bool) {
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

func (f For) HasVariable(name string, oos bool) *Variable {
	for _, v := range f.Vars {
		if v.Name == name {
			return v
		}
	}

	if oos {
		if v := f.definesVariable(name); v != nil {
			return v
		}
	}

	if f.parent != nil {
		return f.parent.HasVariable(name, oos)
	}
	return nil
}

func (f *For) DefineVariable(v *Variable) {
	for _, vr := range f.Vars {
		if vr.Name != v.Name || vr.typ.Eq(v.typ) {
			continue
		}
		vr.typ = NewTyp(Anything, false)
		a, ok := vr.FirstDefinition.(*Assign)
		if !ok {
			panic(`For cycle cannot move to VarDef.`)
		}
		a.FirstDefinition = false
		return
	}
	f.Vars = append(f.Vars, v)
}

func (f *For) definesVariable(name string) *Variable {
	for i, v := range f.Vars {
		if v.Name == name {
			f.unset(i)
			return v
		}
	}
	return f.Block.definesVariable(name)
}

func (f *For) unset(index int) {
	copy(f.Vars[index:], f.Vars[index+1:])
	f.Vars = f.Vars[:len(f.Vars)-1]
	if a, ok := f.Init.(*Assign); ok {
		a.FirstDefinition = false
	}
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
	if len(f.Labels) > 0 {
		s.WriteByte('\n')
	}
	for _, l := range f.Labels {
		s.WriteString(fmt.Sprintf("%s:\n", l.String()))
	}
	return s.String()
}

func NewFor(parent Block) *For {
	f := &For{
		parent: parent,
		Vars:   make([]*Variable, 0),
		Block: &Code{
			Vars:       make([]*Variable, 0),
			Statements: make([]Node, 0),
		},
		Labels: make([]Const, 0),
	}
	f.Block = NewCode(f)
	f.SetParent(parent)
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

	Labels []Const
}

func (f Foreach) Parent() Node {
	return f.parent
}

func (f *Foreach) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	f.parent = n.(Block)
}

func (f Foreach) HasVariable(name string, oos bool) *Variable {
	v := f.definesVariable(name)
	if v != nil {
		return v
	}

	if f.parent != nil {
		return f.parent.HasVariable(name, oos)
	}
	return nil
}

func (f *Foreach) AddStatement(n Node)        {}
func (f *Foreach) DefineVariable(v *Variable) {}

func (f Foreach) definesVariable(name string) *Variable {
	if f.Key != nil && f.Key.Name == name {
		return f.Key
	}
	if f.Value.Name == name {
		return &f.Value
	}
	return nil
}

func (_ *Foreach) unset(index int) {}

func NewForeach(parent Block) *Foreach {
	f := &Foreach{
		parent: parent,
		Labels: make([]Const, 0),
	}
	f.Block = NewCode(f)
	f.SetParent(parent)
	return f
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
	if len(f.Labels) > 0 {
		s.WriteByte('\n')
	}
	for _, l := range f.Labels {
		s.WriteString(fmt.Sprintf("%s:\n", l.String()))
	}
	return s.String()
}

type Switch struct {
	parent Block

	Condition Expression
	// Default will end up here too,
	// to keep the order from PHP.
	Cases []Node

	Labels []Const
}

func (sw Switch) Parent() Node {
	return sw.parent
}

func (sw *Switch) SetParent(n Node) {
	// TODO: Make sure everybody knows
	// it can fail.
	sw.parent = n.(Block)
}

func (sw Switch) HasVariable(name string, oos bool) *Variable {
	if sw.parent != nil {
		return sw.parent.HasVariable(name, oos)
	}
	return nil
}

func (sw *Switch) AddStatement(n Node)        {}
func (sw *Switch) DefineVariable(v *Variable) {}

func (sw Switch) definesVariable(name string) *Variable {
	return nil
}

func (sw Switch) unset(index int) {}

func (sw Switch) String() string {
	s := strings.Builder{}
	s.WriteString("switch ")
	s.WriteString(sw.Condition.String())
	s.WriteString(" {\n")
	for _, c := range sw.Cases {
		s.WriteString(c.String())
	}
	s.WriteByte('}')
	if len(sw.Labels) > 0 {
		s.WriteByte('\n')
	}
	for _, l := range sw.Labels {
		s.WriteString(fmt.Sprintf("%s:\n", l.String()))
	}
	return s.String()
}

type Case struct {
	parent *Switch

	Block     *Code
	Condition Expression
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

func (c *Case) HasVariable(name string, oos bool) *Variable {
	if p := c.parent; p != nil {
		if v := p.HasVariable(name, oos); v != nil {
			return v
		}
	}
	return nil
}

func (c Case) String() string {
	s := strings.Builder{}
	s.WriteString("case ")
	s.WriteString(c.Condition.String())
	s.WriteString(":\n")
	s.WriteString(c.Block.String())
	s.WriteByte('\n')
	return s.String()
}

func (Case) AddStatement(Node)                {}
func (Case) DefineVariable(*Variable)         {}
func (Case) definesVariable(string) *Variable { return nil }
func (Case) unset(int)                        {}

func NewCase(parent *Switch) *Case {
	c := &Case{parent: parent}
	c.Block = NewCode(c)
	c.Block.withBrackets = false
	return c
}

type Default struct {
	parent *Switch

	Block *Code
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

func (d *Default) HasVariable(name string, oos bool) *Variable {
	if p := d.parent; p != nil {
		if v := p.HasVariable(name, oos); v != nil {
			return v
		}
	}
	return nil
}

func (Default) AddStatement(Node)                {}
func (Default) DefineVariable(*Variable)         {}
func (Default) definesVariable(string) *Variable { return nil }
func (Default) unset(int)                        {}

func (d Default) String() string {
	s := strings.Builder{}
	s.WriteString("default:\n")
	s.WriteString(d.Block.String())
	s.WriteByte('\n')
	return s.String()
}

func NewDefault(parent *Switch) *Default {
	d := &Default{parent: parent}
	d.Block = NewCode(d)
	d.Block.withBrackets = false
	return d
}

type If struct {
	parent Block

	Vars []*Variable

	Init Node
	cond Expression

	True  *Code
	False Block
}

func (i *If) SetCond(e Expression) error {
	if !e.Type().Equal(Bool) {
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

func (i If) HasVariable(name string, oos bool) *Variable {
	v := i.definesVariable(name)
	if v != nil {
		return v
	}
	if i.parent != nil {
		return i.parent.HasVariable(name, oos)
	}
	return nil
}

func (i *If) DefineVariable(v *Variable) {
	for _, vr := range i.Vars {
		if vr.Name != v.Name || vr.Type().Eq(v.typ) {
			continue
		}
		vr.typ = NewTyp(Anything, false)
		a, ok := vr.FirstDefinition.(*Assign)
		if !ok {
			panic(`For cycle cannot move to VarDef.`)
		}
		a.FirstDefinition = false
		return
	}
	i.Vars = append(i.Vars, v)
}

func (i *If) definesVariable(name string) *Variable {
	for index, v := range i.Vars {
		if v.Name == name {
			i.unset(index)
			return v
		}
	}
	return nil
}

func (i *If) unset(index int) {
	copy(i.Vars[index:], i.Vars[index+1:])
	i.Vars = i.Vars[:len(i.Vars)-1]
	if a, ok := i.Init.(*Assign); ok {
		a.FirstDefinition = false
	}
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

func NewIf(parent Block) *If {
	return &If{
		parent: parent,
	}
}
