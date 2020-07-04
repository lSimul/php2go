// Package test brings z7zmey/php-parser nodes
// to the p package.
package test

import (
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"
)

// Nop creates empty statement, in a script it is
// represented by semicolon or empty <?php ?>.
func Nop() *stmt.Nop {
	return &stmt.Nop{}
}

// List creates statement list from array, commonly
// represented by a block.
func List(nodes []node.Node) *stmt.StmtList {
	return stmt.NewStmtList(nodes)
}

// Variable creates a new variable with given name.
func Variable(name string) *expr.Variable {
	return &expr.Variable{
		VarName: &node.Identifier{
			Value: name,
		},
	}
}

// String creates simple scalar from given value.
func String(value string) *scalar.String {
	return &scalar.String{
		Value: value,
	}
}

// Func creates empty function with given name.
func Func(name string) *stmt.Function {
	n := &node.Identifier{
		Value: name,
	}
	return stmt.NewFunction(n, false, []node.Node{}, nil, []node.Node{}, "")
}

// Name turns array of string to a name node.
// This is used around constants and return types,
// for instance.
func Name(parts ...string) *name.Name {
	nodes := make([]node.Node, 0)
	for _, n := range parts {
		nodes = append(nodes, &name.NamePart{
			Value: n,
		})
	}
	return name.NewName(nodes)
}

// BinaryOp creates from two nodes binary operand
// specified by op.
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

// Int returns simple data type int with given value.
func Int(value string) *scalar.Lnumber {
	return scalar.NewLnumber(value)
}

// Float returns simple data type float with given value.
func Float(value string) *scalar.Dnumber {
	return scalar.NewDnumber(value)
}

// Plus prepends plus for node e.
func Plus(e node.Node) *expr.UnaryPlus {
	return expr.NewUnaryPlus(e)
}

// Minus prepends minus for node e.
func Minus(e node.Node) *expr.UnaryMinus {
	return expr.NewUnaryMinus(e)
}

// HTML returns node representing HTML outside <?php ?>.
func HTML(value string) *stmt.InlineHtml {
	return stmt.NewInlineHtml(value)
}
