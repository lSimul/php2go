package p

import (
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
		gc.Add(f)
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

func createFunction(b lang.Block, stmts []node.Node) {
	var n lang.Node
	for _, s := range stmts {
		switch s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			n = &lang.HTML{
				Content: s.(*stmt.InlineHtml).Value,
			}
			b.AddStatement(n)

		case *stmt.StmtList:
			list := &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			list.SetParent(b)
			b.AddStatement(list)
			createFunction(list, s.(*stmt.StmtList).Stmts)

		case *stmt.Expression:
			defineExpression(b, s.(*stmt.Expression))

		case *stmt.For:
			f := s.(*stmt.For)
			// TODO: Move definition into the function, this is kinda long.
			lf := &lang.For{
				Vars: make([]lang.Variable, 0),
			}
			lf.SetParent(b)

			if f.Init != nil {
				n := f.Init[0]
				ex := simpleExpression(lf, n)
				lf.Init = ex
			}

			if f.Cond != nil {
				n := f.Cond[0]
				ex := simpleExpression(lf, n)
				lf.Cond = ex
			}

			if f.Loop != nil {
				n := f.Loop[0]
				ex := simpleExpression(lf, n)
				lf.Loop = ex
			}

			lf.Block = &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			lf.Block.SetParent(lf)
			createFunction(lf.Block, []node.Node{f.Stmt})
			b.AddStatement(lf)

		case *stmt.Return:
			r := &lang.Return{
				Expression: complexExpression(b, s.(*stmt.Return).Expr),
			}
			b.AddStatement(r)

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
				f.AddArg(complexExpression(b, e))
			}
			b.AddStatement(f)

		default:
			panic(`Unexpected statement.`)
		}
	}
}

func defineExpression(b lang.Block, e *stmt.Expression) {
	ex := complexExpression(b, e.Expr)
	b.AddStatement(ex)
}

func simpleExpression(b lang.Block, n node.Node) lang.Expression {
	e := expression(b, n)
	if e != nil {
		return e
	}

	switch n.(type) {
	case *assign.Assign:
		a := n.(*assign.Assign)

		r := simpleExpression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		_, ok := r.(*lang.Assign)
		if ok {
			panic(`Simple expression does not support multiple assignments.`)
		}

		n := identifierName(a.Variable.(*expr.Variable))

		var as *lang.Assign
		if v := b.HasVariable(n); v == nil {
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

			b.DefineVariable(*v)
		} else {
			if v.GetType() != r.GetType() {
				panic("Invalid assignment, \"" + v.GetType() + "\" expected, \"" + r.GetType() + "\" given.")
			}

			as = lang.CreateAssign(v, r)
		}

		return as
	}
	panic(`Something else uncatched.`)
}

func complexExpression(b lang.Block, n node.Node) lang.Expression {
	e := expression(b, n)
	if e != nil {
		return e
	}

	switch n.(type) {
	// Every expression should have return value.
	// Otherwise I cannot say what the assigned value will have.
	case *assign.Assign:
		a := n.(*assign.Assign)

		r := expression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		la, ok := r.(*lang.Assign)
		if ok {
			b.AddStatement(la)
			r = la.Left()
		}

		n := identifierName(a.Variable.(*expr.Variable))

		var as *lang.Assign
		if v := b.HasVariable(n); v == nil {
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

			b.DefineVariable(*v)
		} else {
			if v.GetType() != r.GetType() {
				panic("Invalid assignment, \"" + v.GetType() + "\" expected, \"" + r.GetType() + "\" given.")
			}

			as = lang.CreateAssign(v, r)
		}

		return as
	}
	panic(`Something else uncatched.`)
}

func expression(b lang.Block, n node.Node) lang.Expression {
	switch n.(type) {
	case *expr.Variable:
		name := identifierName(n.(*expr.Variable))
		if v := b.HasVariable(name); v == nil {
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
		a := n.(*assign.Assign)

		r := expression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		la, ok := r.(*lang.Assign)
		if ok {
			b.AddStatement(la)
			r = la.Left()
		}

		n := identifierName(a.Variable.(*expr.Variable))

		var as *lang.Assign
		if v := b.HasVariable(n); v == nil {
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

			b.DefineVariable(*v)
		} else {
			if v.GetType() != r.GetType() {
				panic("Invalid assignment, \"" + v.GetType() + "\" expected, \"" + r.GetType() + "\" given.")
			}

			as = lang.CreateAssign(v, r)
		}

		return as

	case *expr.UnaryPlus:
		e := expression(b, n.(*expr.UnaryPlus).Expr)
		e.SetParent(b)
		return e

	case *expr.UnaryMinus:
		m := &lang.UnaryMinus{
			Right: expression(b, n.(*expr.UnaryMinus).Expr),
		}
		m.SetParent(b)
		return m

	case *expr.PostInc:
		i := &lang.Inc{
			Var: expression(b, n.(*expr.PostInc).Variable).(*lang.Variable),
		}
		i.SetParent(b)
		return i

	case *scalar.Lnumber:
		n := &lang.Number{
			Value: n.(*scalar.Lnumber).Value,
		}
		n.SetParent(b)
		return n

	case *scalar.Dnumber:
		s := n.(*scalar.Dnumber).Value
		f := &lang.Float{
			Value: s,
		}
		f.SetParent(b)
		return f

	case *scalar.String:
		s := n.(*scalar.String).Value
		str := &lang.Str{
			Value: s,
		}
		str.SetParent(b)
		return str

	case *binary.Plus:
		p := n.(*binary.Plus)
		op := lang.CreateBinaryOp("+", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.Minus:
		p := n.(*binary.Minus)
		op := lang.CreateBinaryOp("-", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.Mul:
		p := n.(*binary.Mul)
		op := lang.CreateBinaryOp("*", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.Smaller:
		p := n.(*binary.Smaller)
		op := lang.CreateBinaryOp("<", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *expr.FunctionCall:
		fc := n.(*expr.FunctionCall)

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
			f.AddArg(expression(b, a.(*node.Argument).Expr))
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
