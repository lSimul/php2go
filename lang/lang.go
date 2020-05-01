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

	HasVariable(string) *Variable
	definesVariable(string) *Variable
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
