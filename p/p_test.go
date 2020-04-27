package p

import (
	"strings"
	"testing"

	"php2go/lang"
	"php2go/p/test"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/php7"
)

func TestP(t *testing.T) {
	t.Run("helper functions", helpers)
	t.Run("basic set", functionDef)
	t.Run("binary operations", testBinaryOp)
	t.Run("unary operations", unaryOp)
	t.Run("statements", testStatements)
	t.Run("text comparision of the main function", testMain)
}

func helpers(t *testing.T) {
	t.Helper()

	parser := parser{
		translator:         NewNameTranslator(),
		functionTranslator: NewFunctionTranslator(),
		useGlobalContext:   false,
	}

	functions := []struct {
		source   *name.Name
		expected string
	}{
		{test.Name("f"), "f"},
		{test.Name("function"), "function"},
		{test.Name("func"), "func1"},
	}
	for _, f := range functions {
		if name := parser.constructName(f.source, true); name != f.expected {
			t.Errorf("'%s' expected, '%s' found.\n", f.expected, name)
		}
	}

	variables := []struct {
		source   *expr.Variable
		expected string
	}{
		{test.Variable("f"), "f"},
		{test.Variable("func"), "func1"},
		{test.Variable("function"), "function"},
	}
	for _, v := range variables {
		if name := parser.identifierName(v.source); name != v.expected {
			t.Errorf("'%s' expected, '%s' found.\n", v.expected, name)
		}
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

	parser := parser{
		translator:         NewNameTranslator(),
		functionTranslator: NewFunctionTranslator(),
		useGlobalContext:   false,
	}

	// This tests which name and return type will
	// be used. lang.NewFunc(string) is tested
	// elsewhere.
	f := parser.funcDef(nil)
	if f != nil {
		t.Error("From nil nothing can be created.")
	}

	funcDefs := []struct {
		f    *stmt.Function
		name string
		ret  string
	}{
		{test.Func("f"), "f", lang.Void},
		{test.Func("function"), "function", lang.Void},
		{test.Func("func"), "func1", lang.Void},
	}

	for _, f := range funcDefs {
		def := parser.funcDef(f.f)
		if def.Name != f.name {
			t.Errorf("'%s' expected, '%s' found.\n", f.name, def.Name)
		}
		if def.Return != f.ret {
			t.Errorf("'%s' expected, '%s' found.\n", f.name, def.Return)
		}
	}

	f = parser.mainDef()
	if f.Name != "main" {
		t.Errorf("'%s' expected, '%s' found.\n", "main", f.Name)
	}
	if f.Return != lang.Void {
		t.Errorf("'%s' expected, '%s' found.\n", lang.Void, f.Return)
	}

	// It used to be empty string, but because
	// funcDef translates the function name,
	// it had to be changed to something
	// meaningful.
	placeholderFunction := test.Func("placeholderFunction")
	returnTypes := []struct {
		typ      *name.Name
		expected string
	}{
		{test.Name("void"), lang.Void},
		{test.Name("int"), lang.Int},
		{test.Name("string"), lang.String},
	}
	for _, rt := range returnTypes {
		placeholderFunction.ReturnType = rt.typ
		f := parser.funcDef(placeholderFunction)
		if f.Return != rt.expected {
			t.Errorf("'%s' expected, '%s' found.\n", rt.expected, f.Return)

		}
	}
}

func testBinaryOp(t *testing.T) {
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

	parser := parser{}
	for _, c := range cases {
		expr := parser.expression(nil, test.BinaryOp(left, c.op, right))
		op, ok := expr.(*lang.BinaryOp)
		if !ok {
			t.Fatal("Expected binary operation, something else found.")
		}
		if op.Operation != c.op {
			t.Errorf("'%s' expected, '%s' found.", c.op, op.Operation)
		}
		if op.Type() != c.ret {
			t.Errorf("'%s' expected, '%s' found.", c.ret, op.Type())
		}
	}
}

func unaryOp(t *testing.T) {
	t.Helper()

	parent := lang.NewCode(nil)

	parser := parser{}
	for _, n := range []node.Node{
		test.Plus(test.String(`"test"`)),
		test.Plus(test.String(`""`)),
	} {
		e := parser.expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Str); !ok {
			t.Error("lang.Str expected.")
		}
		if typ := e.Type(); typ != lang.String {
			t.Errorf("'string' expected, '%s' found.", typ)
		}
	}

	for _, n := range []node.Node{
		test.Plus(test.Int("0")),
		test.Plus(test.Int("2")),
	} {
		e := parser.expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Number); !ok {
			t.Error("lang.Number expected.")
		}
		if typ := e.Type(); typ != lang.Int {
			t.Errorf("'int' expected, '%s' found.", typ)
		}
	}

	for _, n := range []node.Node{
		test.Plus(test.Float("0")),
		test.Plus(test.Float("1.0")),
	} {
		e := parser.expression(parent, n)
		if e.Parent() != parent {
			t.Error("Parent not set.")
		}
		if _, ok := e.(*lang.Float); !ok {
			t.Error("lang.Float expected.")
		}
		if typ := e.Type(); typ != lang.Float64 {
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
		e := parser.expression(parent, c.n)
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
		if typ := u.Type(); typ != c.t {
			t.Errorf("'%s' expected, '%s' found.", c.t, typ)
		}
	}
}

func testStatements(t *testing.T) {
	t.Helper()

	parser := parser{}

	b := lang.NewCode(nil)
	html := test.HTML("<html></html>")
	parser.createFunction(b, []node.Node{html})
	if len(b.Statements) != 1 {
		t.Fatal("Wrong amount of statements in the block.")
	}
	h, ok := b.Statements[0].(*lang.FunctionCall)
	if !ok {
		t.Fatal("That one statement should be function call.")
	}
	if h.Parent() != b {
		t.Error("Parent not set.")
	}
	if h.Return != lang.Void {
		t.Errorf("'void' expected, '%s' found.", h.Return)
	}
	if h.Name != "fmt.Print" {
		t.Errorf("'fmt.Print' expected, '%s' found.", h.Name)
	}
	if len(h.Args) != 1 {
		t.Fatal("'fmt.Print' should have only one argument.")
	}
	a, ok := h.Args[0].(*lang.Str)
	if !ok {
		t.Fatal("That one argument should be string.")
	}
	if a.Parent() != h {
		t.Error("Parent not set.")
	}
	if a.Value != "`<html></html>`" {
		t.Errorf("'`<html></html>`' expected, '%s' found", a.Value)
	}
	if a.Type() != lang.String {
		t.Errorf("'string' expected, '%s' found.", a.Type())
	}
}

func testMain(tt *testing.T) {
	tt.Helper()

	tests := []struct {
		source   []byte
		expected string
	}{
		// Sandbox
		{
			source: []byte(`<?php
			$a = 1 + 2;
			`),
			expected: `func main() {
A = 1 + 2
}
`,
		},
		// examples/4.php
		{
			source: []byte(`<?php

			$a = 2 + 3 + 4 * 2;

			echo $a * $a;
		`),
			expected: `func main() {
A = 2 + 3 + 4 * 2
fmt.Print(A * A)
}
`,
		},
		// examples/5.php
		// TODO: Get rid of these type casts, define variable only on the lowest level,
		// then check if it is necessary.
		{
			source: []byte(`<?php
			{
				{
					$a = "0";
					// Added to compile it in Go. This var is not used.
					echo $a;
				}
				$a = 1;

				echo $a;
			}
			`),
			expected: `func main() {
{
{
A = "0"
fmt.Print(A.(string))
}
A = 1
fmt.Print(A.(int))
}
}
`,
		},
		// examples/7.php
		{
			source: []byte(`<?php
			$a = 0;
			{
				$a = "1";
				echo $a;
			}
			echo $a;
			$a = 2;
			echo $a;
			`),
			expected: `func main() {
A = 0
{
A = "1"
fmt.Print(A.(string))
}
fmt.Print(A.(string))
A = 2
fmt.Print(A.(int))
}
`,
		},
	}

	for _, t := range tests {
		parser := parser{
			translator:         NewNameTranslator(),
			functionTranslator: NewFunctionTranslator(),
			useGlobalContext:   false,
		}

		out := parser.Run(parsePHP(t.source))
		main := out.Funcs["main"].String()
		if (strings.Compare(main, t.expected)) != 0 {
			tt.Errorf("Expected:\n%s\n Found:\n%s\n", t.expected, main)
		}
	}
}

func parsePHP(source []byte) *node.Root {
	parser := php7.NewParser(source, "")
	parser.Parse()
	return parser.GetRootNode().(*node.Root)
}
