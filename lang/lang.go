package lang

type Node interface {
	HasVariable(string) bool
	SetParent(Node)
	Parent() Node // Might be extra.

	Print()
}

type Block interface {
	HasVariable(string) bool
	Parent() Node

	AddStatement(Node)
}

type Variable struct {
	Type      string
	Name      string
	Const     bool
	Reference bool
}
