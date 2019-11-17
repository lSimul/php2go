package p

import (
	"fmt"
	"reflect"

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

// func createFunction(l *lang.Function, stmts []node.Node) {
func createFunction(l *lang.Function, stmts []node.Node) {
	for _, s := range stmts {
		switch s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?>, nothing to do.

		case *stmt.InlineHtml:
			html := &lang.HTML{
				// Byl jsi krátkozraký, blok může patřit funkci, spousta věcí může patřit asi jen funkci.
				// A blok může patřit bloku atd., hnusné to je.
				//
				// K tomuto převádění nebude docházet, bude to všechno krásně k bloku. Vždycky.
				// Nebude problém ani s funkcí, ty argumenty se hodí jako proměnná do bloku, možná lehce zprzněné
				// o příznak zda už bylo definováno, něco takového. Ani to asi nebude třeba.
				// Proměnné takhle umístěné slouží jedinému účelu, zkoumání zda je něco nedefinovaného,
				// definováno nějak hnusně mimo scope atp.
				// Možná budu muset přidat další chytrosti a zas tak přímočaře to nepojede,
				//
				// for ($i = 0; $i < 10; $i++) {} echo $i;
				//
				// Je bohužel validní PHP kód co vypíše 10, to teď se svoji strukturou nezvládnu vyřešit.
				Parent:  l.Body,
				Content: s.(*stmt.InlineHtml).Value,
			}
			l.Body.Statements = append(l.Body.Statements, html)
		default:
			panic(`Unexpected statement.`)
		}
	}
}

func ptr(obj *lang.Function) *lang.Node {
	vp := reflect.New(reflect.TypeOf(obj))
	vp.Elem().Set(reflect.ValueOf(obj))
	return vp.Interface().(*lang.Node)
}
