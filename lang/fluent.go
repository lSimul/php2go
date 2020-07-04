package lang

import (
	"fmt"
	"strings"
)

type direment struct {
	parent Node

	strV func(Expression) (Expression, error)
	v    *VarRef
}

func (d direment) Parent() Node {
	return d.parent
}

func (d *direment) SetParent(n Node) {
	d.parent = n
}

func (d direment) UsedVar() *Variable {
	return d.v.V
}

func (d direment) str(unary, binary string) string {
	s := strings.Builder{}

	if d.v.typ.Equal(String) {
		if d.strV == nil {
			s.WriteString("/* undefined stringIncrement */")
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
		b, _ := NewBinaryOp(binary, NewVarRef(d.v.V, d.v.typ), &Number{Value: "1"})
		a, _ := NewAssign(d.v.V, b)
		s.WriteString(a.String())
	} else {
		if d.v.typ.IsPointer {
			s.WriteByte('*')
		}
		s.WriteString(d.v.String())
		s.WriteString(unary)
	}
	return s.String()
}

type Inc struct {
	direment
}

func NewInc(parent Node, v *VarRef, strV func(Expression) (Expression, error)) *Inc {
	return &Inc{
		direment: direment{
			parent: parent,
			v:      v,
			strV:   strV,
		},
	}
}

func (i Inc) String() string {
	return i.str("++", "+")
}

type Dec struct {
	direment
}

func NewDec(parent Node, v *VarRef, strV func(Expression) (Expression, error)) *Dec {
	return &Dec{
		direment: direment{
			parent: parent,
			v:      v,
			strV:   strV,
		},
	}
}

func (d Dec) String() string {
	return d.str("--", "-")
}
