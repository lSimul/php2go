package lang

import "fmt"

const (
	Void    = ""
	Int     = "int"
	String  = "string"
	Float64 = "float64"
	Bool    = "bool"
)

type Node interface {
	SetParent(Node)
	Parent() Node
	Print()
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

	CurrentType string
}

func (v Variable) Print() {
	fmt.Print(v.Name)
}

func (v Variable) GetType() string {
	return v.CurrentType
}
