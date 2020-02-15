package p

import (
	"testing"

	"php2go/lang"
	"php2go/p/test"

	"github.com/z7zmey/php-parser/node"
)

func TestP(t *testing.T) {
	t.Run("helper functions", helpers)
	t.Run("basic set", functionDef)
	t.Run("binary operations", binaryOp)
	t.Run("unary operations", unaryOp)
}

func helpers(t *testing.T) {
	t.Helper()

	nname := test.Name("f")
	if res := constructName(nname); res != "f" {
		t.Errorf("'%s' expected, '%s' found.\n", "f", res)
	}

	nname = test.Name("function")
	if res := constructName(nname); res != "function" {
		t.Errorf("'%s' expected, '%s' found.\n", "function", res)
	}

	nname = test.Name("func")
	if res := constructName(nname); res != "function" {
		t.Errorf("'%s' expected, '%s' found.\n", "function", res)
	}

	name := "f"
	res := identifierName(test.Variable(name))
	if res != name {
		t.Errorf("'%s' expected, '%s' found.\n", name, res)
	}
	name = "func"
	res = identifierName(test.Variable(name))
	if res != name {
		t.Errorf("'%s' expected, '%s' found.\n", name, res)
	}
	name = "function"
	res = identifierName(test.Variable(name))
	if res != name {
		t.Errorf("'%s' expected, '%s' found.\n", name, res)
	}

	nop := test.Nop()
	if l := nodeList(nop); l[0] != nop {
		t.Error("Nothing should happen to passed node.")
	}
	if l := nodeList(nil); len(l) != 0 {
		t.Error("Nil cannot create non-empty statement list.")
	}
	list := test.List([]node.Node{nop})
	if l := nodeList(list); l[0] != nop {
		t.Error("Nothing should happen to the nodes passed in the node list.")
	}
}

func functionDef(t *testing.T) {
	t.Helper()

	// This tests which name and return type will
	// be used. lang.NewFunc(string) is tested
	// elsewhere.
	f := funcDef(nil)
	if f != nil {
		t.Error("From nil nothing can be created.")
	}

	name := "f"
	nf := test.Func(name)
	f = funcDef(nf)
	if f.Name != name {
		t.Error("Wrong name.")
	}
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	name = "function"
	nf = test.Func(name)
	f = funcDef(nf)
	if f.Name != name {
		t.Error("Wrong name.")
	}
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	name = "func"
	nf = test.Func(name)
	f = funcDef(nf)
	if f.Name != "function" {
		t.Error("Wrong name.")
	}
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	f = mainDef()
	if f.Name != "main" {
		t.Error("Wrong name.")
	}
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	nf = test.Func("")

	nf.ReturnType = test.Name("void")
	f = funcDef(nf)
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	nf.ReturnType = test.Name("void")
	f = funcDef(nf)
	if f.Return != lang.Void {
		t.Error("Wrong return type.")
	}

	nf.ReturnType = test.Name("int")
	f = funcDef(nf)
	if f.Return != lang.Int {
		t.Error("Wrong return type.")
	}

	nf.ReturnType = test.Name("string")
	f = funcDef(nf)
	if f.Return != lang.String {
		t.Error("Wrong return type.")
	}
}

func binaryOp(t *testing.T) {
	t.Helper()

	left := test.Int("1")
	right := test.Int("2")
	cases := []struct {
		op  string
		ret string
	}{
		{"+", lang.Int},
		{"-", lang.Int},
		{"*", lang.Int},
		{"<", lang.Bool},
		{"<=", lang.Bool},
		{">=", lang.Bool},
		{">", lang.Bool},
		{"==", lang.Bool},
	}
	for _, c := range cases {
		expr := expression(nil, test.BinaryOp(left, c.op, right))
		op, ok := expr.(*lang.BinaryOp)
		if !ok {
			t.Fatal("Expected binary operation, something else found.")
		}
		if op.Operation != c.op {
			t.Errorf("'%s' expected, '%s' found.", c.op, op.Operation)
		}
		if op.GetType() != c.ret {
			t.Errorf("'%s' expected, '%s' found.", c.ret, op.GetType())
		}
	}
}

func unaryOp(t *testing.T) {
	t.Helper()

	parent := lang.NewCode(nil)

	for _, n := range []node.Node{
		test.Plus(test.String(`"test"`)),
		test.Plus(test.String(`""`)),
	} {
		e := expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Str); !ok {
			t.Error("lang.Str expected.")
		}
		if typ := e.GetType(); typ != lang.String {
			t.Errorf("'string' expected, '%s' found.", typ)
		}
	}

	for _, n := range []node.Node{
		test.Plus(test.Int("0")),
		test.Plus(test.Int("2")),
	} {
		e := expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Number); !ok {
			t.Error("lang.Number expected.")
		}
		if typ := e.GetType(); typ != lang.Int {
			t.Errorf("'int' expected, '%s' found.", typ)
		}
	}

	for _, n := range []node.Node{
		test.Plus(test.Float("0")),
		test.Plus(test.Float("1.0")),
	} {
		e := expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Float); !ok {
			t.Error("lang.Float expected.")
		}
		if typ := e.GetType(); typ != lang.Float64 {
			t.Errorf("'float' expected, '%s' found.", typ)
		}
	}

	for _, c := range []struct {
		n node.Node
		t string
	}{
		{test.Minus(test.String(`"test"`)), lang.String},
		{test.Minus(test.String(`""`)), lang.String},
		{test.Minus(test.Int("0")), lang.Int},
		{test.Minus(test.Int("2")), lang.Int},
		{test.Minus(test.Float("0")), lang.Float64},
		{test.Minus(test.Float("1.0")), lang.Float64},
	} {
		e := expression(parent, c.n)
		u, ok := e.(*lang.UnaryMinus)
		if !ok {
			t.Fatal("lang.UnaryMinus expected.")
		}
		if u.Parent() != parent {
			t.Error("Parent not set.")
		}
		if u.Expr.Parent() != u {
			t.Error("Parent not set.")
		}
		if typ := u.GetType(); typ != c.t {
			t.Errorf("'%s' expected, '%s' found.", c.t, typ)
		}
	}
}
