package p

import (
	"strconv"

	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/expr"
	"github.com/z7zmey/php-parser/node/expr/assign"
	"github.com/z7zmey/php-parser/node/expr/binary"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/scalar"
	"github.com/z7zmey/php-parser/node/stmt"

	"php2go/lang"
)

var gc *lang.GlobalContext

func Run(r *node.Root) *lang.GlobalContext {
	gc = lang.CreateGlobalContext()
	ms, fs := sanitizeRootStmts(r)

	for _, s := range fs {
		f := funcDef(&s)
		gc.Add(f)
	}
	for _, s := range fs {
		f := funcDef(&s)
		createFunction(&f.Body, s.Stmts)
	}

	main := mainDef()
	gc.Add(main)
	createFunction(&main.Body, ms)
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

func funcDef(fc *stmt.Function) *lang.Function {
	f := lang.CreateFunc(fc.FunctionName.(*node.Identifier).Value)

	for _, pr := range fc.Params {
		p := pr.(*node.Parameter)
		v := lang.Variable{
			Type:      constructName(p.VariableType.(*name.Name)),
			Name:      identifierName(p.Variable.(*expr.Variable)),
			Const:     false,
			Reference: false,
		}
		f.Args = append(f.Args, v)
	}

	if fc.ReturnType != nil {
		n := constructName(fc.ReturnType.(*name.Name))
		if n == "void" {
			n = lang.Void
		}
		f.Return = n
	}

	return f
}

func mainDef() *lang.Function {
	return lang.CreateFunc("main")
}

func createFunction(l *lang.Code, stmts []node.Node) {
	var n lang.Node
	for _, s := range stmts {
		switch s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			n = &lang.HTML{
				Content: s.(*stmt.InlineHtml).Value,
			}
			l.AddStatement(n)

		case *stmt.StmtList:
			list := &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			list.SetParent(l)
			l.AddStatement(list)
			createFunction(list, s.(*stmt.StmtList).Stmts)

		case *stmt.Expression:
			defineExpression(l, s.(*stmt.Expression))

		// Return is not an expression, this is fucked up (for my structure).
		case *stmt.Return:
			r := &lang.Return{
				Expression: expression(l, s.(*stmt.Return).Expr),
			}
			l.AddStatement(r)

		case *stmt.Echo:
			f := &lang.FunctionCall{
				Name: "fmt.Print",
				Args: make([]lang.Expression, 0),
			}

			ex := s.(*stmt.Echo)
			for _, e := range ex.Exprs {
				// TODO: Do not ignore information in Argument,
				// it has interesting information like if it is
				// send by reference and others.
				f.AddArg(expression(l, e))
			}
			l.AddStatement(f)

		default:
			panic(`Unexpected statement.`)
		}
	}
}

func defineExpression(l *lang.Code, e *stmt.Expression) {
	ex := expression(l, e.Expr)
	l.AddStatement(ex)
}

func expression(l *lang.Code, nn node.Node) lang.Expression {
	switch nn.(type) {
	case *expr.Variable:
		name := identifierName(nn.(*expr.Variable))
		if v := (*l).HasVariable(name); v == nil {
			panic("Using undefined variable \"" + name + "\".")
		}
		return &lang.Variable{
			// Type will be taken from the right side.
			Type:      lang.Int,
			Name:      name,
			Const:     false,
			Reference: false,
		}

	// Every expression should have return value.
	// Otherwise I cannot say what the assigned value will have.
	case *assign.Assign:
		a := nn.(*assign.Assign)

		r := expression(l, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		la, ok := r.(*lang.Assign)
		if ok {
			l.AddStatement(la)
			r = la.Left()
		}

		n := identifierName(a.Variable.(*expr.Variable))

		var as *lang.Assign
		if v := l.HasVariable(n); v == nil {
			v = &lang.Variable{
				Type:      r.GetType(),
				Name:      n,
				Const:     false,
				Reference: false,
			}
			if v.Type == lang.Void {
				panic("Cannot assign \"void\" " + "to \"" + n + "\".")
			}
			as = lang.CreateAssign(v, r)
			as.FirstDefinition = true

			l.DefineVariable(*v)
		} else {
			if v.GetType() != r.GetType() {
				panic("Invalid assignment, \"" + v.GetType() + "\" expected, \"" + r.GetType() + "\" given.")
			}

			as = lang.CreateAssign(v, r)
		}

		return as

	// TODO: Set parent
	case *expr.UnaryPlus:
		return expression(l, nn.(*expr.UnaryPlus).Expr)

	// TODO: Set parent
	case *expr.UnaryMinus:
		return &lang.UnaryMinus{
			Right: expression(l, nn.(*expr.UnaryMinus).Expr),
		}

	// TODO: Set parent
	case *expr.PostInc:
		return &lang.Inc{
			Var: expression(l, nn.(*expr.PostInc).Variable).(*lang.Variable),
		}

	// TODO: Set parent
	case *scalar.Lnumber:
		s := nn.(*scalar.Lnumber).Value
		i, _ := strconv.Atoi(s)
		return &lang.Number{
			Value: i,
		}

	// TODO: Set parent
	case *scalar.Dnumber:
		s := nn.(*scalar.Dnumber).Value
		return &lang.Float{
			Value: s,
		}

	// TODO: Set parent
	case *scalar.String:
		s := nn.(*scalar.String).Value
		return &lang.Str{
			Value: s,
		}

	// TODO: Refactor every binary operation, this stinks, alot.

	// TODO: Set parent
	case *binary.Plus:
		p := nn.(*binary.Plus)
		return &lang.BinaryOp{
			Operation: "+",

			Left:  expression(l, p.Left),
			Right: expression(l, p.Right),
		}

	// TODO: Set parent
	case *binary.Minus:
		p := nn.(*binary.Minus)
		return &lang.BinaryOp{
			Operation: "-",

			Left:  expression(l, p.Left),
			Right: expression(l, p.Right),
		}

	// TODO: Set parent
	case *binary.Mul:
		p := nn.(*binary.Mul)
		return &lang.BinaryOp{
			Operation: "*",

			Left:  expression(l, p.Left),
			Right: expression(l, p.Right),
		}

	// TODO: Set parent
	case *expr.FunctionCall:
		fc := nn.(*expr.FunctionCall)

		n := constructName(fc.Function.(*name.Name))
		f := &lang.FunctionCall{
			Name:   n,
			Args:   make([]lang.Expression, 0),
			Return: gc.Get(n).GetType(),
		}

		al := fc.ArgumentList
		for _, a := range al.Arguments {
			// TODO: Do not ignore information in Argument,
			// it has interesting information like if it is
			// send by reference and others.
			f.AddArg(expression(l, a.(*node.Argument).Expr))
		}
		return f

	default:
		panic(`Something else uncatched.`)
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

func constructName(nm *name.Name) string {
	res := ""
	for _, n := range nm.Parts {
		res += n.(*name.NamePart).Value
	}
	return res
}
