package lang

import "fmt"

const (
	Void     = ""
	Int      = "int"
	String   = "string"
	Float64  = "float64"
	Bool     = "bool"
	Anything = "interface{}"
	Writer   = "io.Writer"

	SQL = "std.SQL"
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

	Type() Typ
}

type Typ struct {
	typ       string
	IsPointer bool
	reference bool

	Adressable bool
	Tiles      map[string]Typ
}

func NewTyp(typ string, IsPointer bool) Typ {
	return Typ{typ, IsPointer, false, false, make(map[string]Typ)}
}

func (t Typ) Format() string {
	switch t.typ {
	case Int:
		return "%d"

	case Float64:
		return "%g"

	case String:
		return "%s"

	default:
		return "%v"
	}
}

func (t Typ) String() string {
	if t.Adressable {
		n := "struct{\n"
		for k, v := range t.Tiles {
			n += k + " " + v.String() + "\n"
		}
		n += "}"
		return n
	}

	if t.IsPointer {
		return "*" + t.typ
	}
	return t.typ
}

func (t Typ) Equal(s string) bool {
	return s == t.typ
}

func (t Typ) Eq(r Typ) bool {
	return t.typ == r.typ
}

type Variable struct {
	typ   Typ
	Name  string
	Const bool

	FirstDefinition Node

	CurrentType Typ
}

func (v Variable) String() string {
	return v.Name
}

func (v Variable) Type() Typ {
	return v.CurrentType
}

func NewVariable(name string, typ Typ, isConst bool) *Variable {
	return &Variable{
		Name:  name,
		typ:   typ,
		Const: isConst,

		CurrentType: typ,
	}
}
