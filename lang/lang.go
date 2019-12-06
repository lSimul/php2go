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
	HasVariable(string) *Variable
	SetParent(Node)
	Parent() Node // Might be extra.
	Print()
}

type Block interface {
	HasVariable(string) *Variable
	SetParent(Node)
	Parent() Node
	Print()

	AddStatement(Node)
	DefineVariable(Variable)

	DefinesVariable(string) *Variable
}

type Variable struct {
	Type      string
	Name      string
	Const     bool
	Reference bool
}

func (v *Variable) HasVariable(n string) *Variable {
	if n == v.Name {
		return v
	}
	return nil
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

type Expression interface {
	// Do I need this?
	HasVariable(string) *Variable
	// Uncommented, addStatement did not work because of this.

	SetParent(Node)
	Parent() Node
	Print()

	GetType() string
}
