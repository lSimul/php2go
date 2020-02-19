package test

import (
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
)

func Nop() *stmt.Nop {
	return &stmt.Nop{}
}

func List(nodes []node.Node) *stmt.StmtList {
	return stmt.NewStmtList(nodes)
}

func Variable(name string) *expr.Variable {
	return &expr.Variable{
		VarName: &node.Identifier{
			Value: name,
		},
	}
}

func String(value string) *scalar.String {
	return &scalar.String{
		Value: value,
	}
}

func Func(name string) *stmt.Function {
	n := &node.Identifier{
		Value: name,
	}
	return stmt.NewFunction(n, false, []node.Node{}, nil, []node.Node{}, "")
}

func Name(parts ...string) *name.Name {
	nodes := make([]node.Node, 0)
	for _, n := range parts {
		nodes = append(nodes, &name.NamePart{
			Value: n,
		})
	}
	return name.NewName(nodes)
}

func BinaryOp(left node.Node, op string, right node.Node) node.Node {
	switch op {
	case "+":
		return binary.NewPlus(left, right)
	case "-":
		return binary.NewMinus(left, right)
	case "*":
		return binary.NewMul(left, right)
	case "<":
		return binary.NewSmaller(left, right)
	case "<=":
		return binary.NewSmallerOrEqual(left, right)
	case ">=":
		return binary.NewGreaterOrEqual(left, right)
	case ">":
		return binary.NewGreater(left, right)
	case "==":
		return binary.NewIdentical(left, right)
	}
	return nil
}

func Int(value string) *scalar.Lnumber {
	return scalar.NewLnumber(value)
}

func Float(value string) *scalar.Dnumber {
	return scalar.NewDnumber(value)
}

func Plus(e node.Node) *expr.UnaryPlus {
	return expr.NewUnaryPlus(e)
}

func Minus(e node.Node) *expr.UnaryMinus {
	return expr.NewUnaryMinus(e)
}

func HTML(value string) *stmt.InlineHtml {
	return stmt.NewInlineHtml(value)
}
