package p

import (
	"fmt"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/stmt"

	"php2go/lang"
)

func Run(r *node.Root) *lang.GlobalContext {
	gc := lang.CreateGlobalContext()
	ms, fs := sanitizeRootStmts(r)
	main := mainDef()
	createFunction(main, ms)

	gc.Add(*main)

	for _, s := range fs {
		f := funcDef(&s)
		createFunction(f, s.Stmts)
		fmt.Println(f.Name, s.Stmts)
	}
	return gc
}

// SanitizeRootStmts splits statements based on their type,
// functions go to the one array, rest to the other one.
// This makes function root more straight forward, main will
// not be longer split up by functions.
func sanitizeRootStmts(r *node.Root) ([]node.Node, []stmt.Function) {
	main := make([]node.Node, 0)
	functions := make([]stmt.Function, 0)

	for _, s := range r.Stmts {
		switch s.(type) {
		case *stmt.Function:
			functions = append(functions, *s.(*stmt.Function))
		default:
			main = append(main, s)
		}
	}

	return main, functions
}

func funcDef(f *stmt.Function) *lang.Function {
	fn := f.FunctionName
	return &lang.Function{
		Name: fn.(*node.Identifier).Value,
	}
}

func mainDef() *lang.Function {
	return lang.CreateMain()
}

// func createFunction(l *lang.Function, stmts []node.Node) {
func createFunction(l *lang.Function, stmts []node.Node) {
	var n lang.Node
	for _, s := range stmts {
		switch s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			n = lang.HTML{
				Content: s.(*stmt.InlineHtml).Value,
			}
			l.AddStatement(n)

		case *stmt.Expression:
			defineExpression(l, s.(*stmt.Expression))

		default:
			panic(`Unexpected statement.`)
		}
	}
}

func defineExpression(l *lang.Function, e *stmt.Expression) {
	switch e.Expr.(type) {
	case *expr.Variable:
		name := identifierName(e.Expr.(*expr.Variable))
		if !(*l).HasVariable(name) {
			panic("Using undefined variable \"" + name + "\".")
		}

	// Every expression should have return value.
	// Otherwise I cannot say what the assigned value will have.
	case *assign.Assign:
		a := e.Expr.(*assign.Assign)
		n := identifierName(a.Variable.(*expr.Variable))

		v := lang.Variable{
			// Type will be taken from the right side.
			Type:      "int",
			Name:      n,
			Const:     false,
			Reference: false,
		}
		l.DefineVariable(v)
	}
}

/**
 * Function makes things much easier, I expect
 * identifier name to be just simple right now
 * defined string, no variable etc.
 */
func identifierName(v *expr.Variable) string {
	switch v.VarName.(type) {
	case *node.Identifier:
		return v.VarName.(*node.Identifier).Value

	default:
		panic(`Variable name is not defined as a simple string.`)
	}
}
