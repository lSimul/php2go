package lang

type Node interface {
	HasVariable(string) bool
}

type Function struct {
	Name   string
	Args   map[string]string
	Return string
}

// I moved this function to another interface,
// function will not have parent for now.
/*
func (f Function) ParentNode() *Node {
	return nil
}
*/

func (f Function) HasVariable(name string) bool {
	_, ok := f.Args[name]
	return ok
}

type Statement interface {
	ParentNode() *Node
}

type Block struct {
	parent     *Node
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

func (b Block) ParentNode() *Node {
	return b.parent
}
