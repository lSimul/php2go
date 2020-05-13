package lang

import (
	"fmt"
	"strings"
)

type Inc struct {
	parent Node

	strV func(Expression) (Expression, error)
	v    *VarRef
}

func (i Inc) Parent() Node {
	return i.parent
}

func (i *Inc) SetParent(n Node) {
	i.parent = n
}

func (i Inc) UsedVar() *Variable {
	return i.v.V
}

func (i Inc) String() string {
	s := strings.Builder{}

	if i.v.typ.Equal(String) {
		if i.strV == nil {
			s.WriteString("/* undefined stringIncrement */")
		} else {
			fc, err := i.strV(i.v)
			if err != nil {
				s.WriteString(fmt.Sprintf("/* %v */", err))
			} else {
				fc.SetParent(i.parent)
				a, _ := NewAssign(i.v.V, fc)
				s.WriteString(a.String())
			}
		}
	} else if i.v.V.typ.Equal(Anything) {
		b, _ := NewBinaryOp("+", NewVarRef(i.v.V, i.v.typ), &Number{Value: "1"})
		a, _ := NewAssign(i.v.V, b)
		s.WriteString(a.String())
	} else {
		if i.v.typ.IsPointer {
			s.WriteByte('*')
		}
		s.WriteString(i.v.String())
		s.WriteString("++")
	}
	return s.String()
}

func NewInc(parent Node, v *VarRef, strV func(Expression) (Expression, error)) *Inc {
	return &Inc{
		parent: parent,
		v:      v,
		strV:   strV,
	}
}

type Dec struct {
	parent Node

	strV func(Expression) (Expression, error)
	v    *VarRef
}

func (d Dec) Parent() Node {
	return d.parent
}

func (d *Dec) SetParent(n Node) {
	d.parent = n
}

func (d Dec) String() string {
	s := strings.Builder{}

	if d.v.typ.Equal(String) {
		if d.strV == nil {
			s.WriteString("/* undefined stringDecrement */")
		} else {
			fc, err := d.strV(d.v)
			if err != nil {
				s.WriteString(fmt.Sprintf("/* %v */", err))
			} else {
				fc.SetParent(d.parent)
				a, _ := NewAssign(d.v.V, fc)
				s.WriteString(a.String())
			}
		}
	} else if d.v.V.typ.Equal(Anything) {
		b, _ := NewBinaryOp("-", NewVarRef(d.v.V, d.v.typ), &Number{Value: "1"})
		a, _ := NewAssign(d.v.V, b)
		s.WriteString(a.String())
	} else {
		if d.v.typ.IsPointer {
			s.WriteByte('*')
		}
		s.WriteString(d.v.String())
		s.WriteString("--")
	}
	return s.String()
}

func NewDec(parent Node, v *VarRef, strV func(Expression) (Expression, error)) *Dec {
	return &Dec{
		parent: parent,
		v:      v,
		strV:   strV,
	}
}

func (d Dec) UsedVar() *Variable {
	return d.v.V
}
