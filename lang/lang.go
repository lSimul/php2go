package lang

type Node interface {
	HasVariable(string) bool
}

type Function struct {
	Name   string
	Args   map[string]string
	Return string
	// Statement is required to hack p.go createFunction type conversion
	Body *Statement
}

// I moved this function to another interface,
// function will not have parent for now.
/*
func (f Function) ParentNode() *Statement {
	return nil
}
*/

func (f Function) HasVariable(name string) bool {
	_, ok := f.Args[name]
	return ok
}

type Statement interface {
	HasVariable(string) bool
	ParentNode() *Statement
}

type Block struct {
	parent     *Statement
	Vars       map[string]string
	Statements []Statement
}

func (b Block) HasVariable(name string) bool {
	_, ok := b.Vars[name]
	if ok {
		return ok
	}

	if b.parent != nil {
		return (*b.parent).HasVariable(name)
	}
	return false
}

func (b Block) ParentNode() *Statement {
	return b.parent
}

// Refactor to something like function call
// function call fmt.Println or something
// different what should be in the standard
// library.
type HTML struct {
	Parent  *Statement
	Content string
}

func (h HTML) ParentNode() *Statement {
	return h.Parent
}

func (h HTML) HasVariable(name string) bool {
	if h.Parent != nil {
		return (*h.Parent).HasVariable(name)
	}
	return false
}
