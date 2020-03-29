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
	DefinesVariable(string) *Variable
}

type Expression interface {
	Node

	GetType() string
}

type Variable struct {
	Type  string
	Name  string
	Const bool

	FirstDefinition Node

	CurrentType string
}

func (v Variable) String() string {
	return v.Name
}

func (v Variable) GetType() string {
	return v.CurrentType
}
