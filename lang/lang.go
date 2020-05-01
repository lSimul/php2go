package lang

import "fmt"

const (
	Void     = ""
	Int      = "int"
	String   = "string"
	Float64  = "float64"
	Bool     = "bool"
	Anything = "interface{}"
)

type Node interface {
	fmt.Stringer

	SetParent(Node)
	Parent() Node
}

type Block interface {
	Node

	AddStatement(Node)
	DefineVariable(*Variable)

	// HasVariable tries to find a variable definition
	// at any cost, even searching out of scope. It is
	// bounded to the DefinesVariable, this function
	// does the out of scope search, HasVariable changes
	// the definition so it can be used in Go.
	HasVariable(string, bool) *Variable
	definesVariable(string) *Variable
	unset(index int)
}

type Expression interface {
	Node

	Type() string
}

type Variable struct {
	typ   string
	Name  string
	Const bool

	FirstDefinition Node

	CurrentType string
}

func (v Variable) String() string {
	return v.Name
}

func (v Variable) Type() string {
	return v.CurrentType
}

func NewVariable(name, typ string, isConst bool) *Variable {
	return &Variable{
		Name:  name,
		typ:   typ,
		Const: isConst,

		CurrentType: typ,
	}
}
