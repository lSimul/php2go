package p

import (
	"fmt"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/stmt"

	"php2go/lang"
)

func Run(r *node.Root) {

	ms, fs := sanitizeRootStmts(r)
	main := mainDef()
	createFunction(main, ms)

	for _, s := range fs {
		f := funcDef(&s)
		createFunction(f, s.Stmts)
		fmt.Println(f.Name, s.Stmts)
	}
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
	return &lang.Function{
		Name: "main",
	}
}

func createFunction(l *lang.Function, stmts []node.Node) {
}
