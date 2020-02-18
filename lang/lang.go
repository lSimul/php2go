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
	SetParent(Node)
	Parent() Node
	Print()

	AddStatement(Node)
	DefineVariable(Variable)

	HasVariable(string) *Variable
	DefinesVariable(string) *Variable
}

type Expression interface {
	SetParent(Node)
	Parent() Node
	Print()

	GetType() string
}

type Variable struct {
	Type      string
	Name      string
	Const     bool
	Reference bool
}

func (v Variable) SetParent(Node) {}

func (v Variable) Parent() Node {
	return nil
}

func (v Variable) Print() {
	fmt.Print(v.Name)
}

func (v Variable) GetType() string {
	return v.Type
}
