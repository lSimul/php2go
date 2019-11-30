package p

import (
	"errors"

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
	n := fc.FunctionName.(*node.Identifier).Value
	if n == "func" {
		n = "function"
	}
	f := lang.CreateFunc(n)

	for _, pr := range fc.Params {
		p := pr.(*node.Parameter)
		v := lang.Variable{
			Type:      constructName(p.VariableType.(*name.Name)),
			Name:      identifierName(p.Variable.(*expr.Variable)),
			Const:     false,
			Reference: p.ByRef,
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

		// TODO: Move definition into the function, this is kinda long.
		case *stmt.For:
			f := s.(*stmt.For)
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

			list, ok := f.Stmt.(*stmt.StmtList)
			if ok {
				createFunction(lf.Block, list.Stmts)
			} else {
				createFunction(lf.Block, []node.Node{f.Stmt})
			}
			b.AddStatement(lf)

		case *stmt.While:
			w := s.(*stmt.While)
			f := &lang.For{
				Vars: make([]lang.Variable, 0),
			}
			f.SetParent(b)

			ex := simpleExpression(f, w.Cond)
			f.Cond = ex

			f.Block = &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			f.Block.SetParent(f)

			list, ok := w.Stmt.(*stmt.StmtList)
			if ok {
				createFunction(f.Block, list.Stmts)
			} else {
				createFunction(f.Block, []node.Node{w.Stmt})
			}
			b.AddStatement(f)

		case *stmt.Do:
			w := s.(*stmt.Do)
			lf := &lang.For{
				Vars: make([]lang.Variable, 0),
			}
			lf.SetParent(b)

			lf.Block = &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			lf.Block.SetParent(lf)

			list, ok := w.Stmt.(*stmt.StmtList)
			if ok {
				createFunction(lf.Block, list.Stmts)
			} else {
				createFunction(lf.Block, []node.Node{w.Stmt})
			}

			i := &lang.If{
				Vars: make([]lang.Variable, 0),
				True: &lang.Code{
					Vars:       make([]lang.Variable, 0),
					Statements: make([]lang.Node, 0),
				},
			}
			i.True.SetParent(i)
			i.SetParent(lf)
			c := simpleExpression(i, w.Cond)
			neg := &lang.Negation{
				Right: c,
			}
			c.SetParent(neg)
			i.Cond = neg
			neg.SetParent(i)
			i.True.AddStatement(&lang.Break{})
			lf.Block.AddStatement(i)

			b.AddStatement(lf)
		//

		case *stmt.If:
			i := constructIf(b, s.(*stmt.If))
			b.AddStatement(i)

		case *stmt.Switch:
			s := s.(*stmt.Switch)
			sw := &lang.Switch{
				Cases: make([]lang.Node, 0),
			}
			sw.SetParent(b)
			sw.Condition = expression(sw, s.Cond)
			constructSwitch(sw, s.CaseList)
			b.AddStatement(sw)

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

		case *stmt.Break:
			br := &lang.Break{}
			br.SetParent(b)
			b.AddStatement(br)

		case *stmt.Continue:
			c := &lang.Continue{}
			c.SetParent(b)
			b.AddStatement(c)

		default:
			panic(`Unexpected statement.`)
		}
	}
}

func constructIf(b lang.Node, i *stmt.If) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	nif.Cond = simpleExpression(nif, i.Cond)
	nif.True = &lang.Code{
		Vars:       make([]lang.Variable, 0),
		Statements: make([]lang.Node, 0),
	}
	nif.True.SetParent(nif)
	list, ok := i.Stmt.(*stmt.StmtList)
	if ok {
		createFunction(nif.True, list.Stmts)
	} else {
		createFunction(nif.True, []node.Node{i.Stmt})
	}

	lif := nif
	for _, ei := range i.ElseIf {
		lif.False = constructElif(b, ei.(*stmt.ElseIf))
		lif = lif.False.(*lang.If)
	}

	if i.Else != nil {
		e := i.Else.(*stmt.Else).Stmt
		switch e.(type) {
		case *stmt.If:
			lif.False = constructIf(lif, e.(*stmt.If))

		default:
			c := &lang.Code{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			c.SetParent(lif)
			list, ok := e.(*stmt.StmtList)
			if ok {
				createFunction(c, list.Stmts)
			} else {
				createFunction(c, []node.Node{e})
			}
			lif.False = c
		}
	}
	return nif
}

func constructElif(b lang.Node, i *stmt.ElseIf) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	nif.Cond = simpleExpression(nif, i.Cond)
	nif.True = &lang.Code{
		Vars:       make([]lang.Variable, 0),
		Statements: make([]lang.Node, 0),
	}
	nif.True.SetParent(nif)
	list, ok := i.Stmt.(*stmt.StmtList)
	if ok {
		createFunction(nif.True, list.Stmts)
	} else {
		createFunction(nif.True, []node.Node{i.Stmt})
	}
	return nif
}

func constructSwitch(s *lang.Switch, cl *stmt.CaseList) {
	for _, c := range cl.Cases {
		switch c.(type) {
		case *stmt.Case:
			c := c.(*stmt.Case)
			lc := &lang.Case{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			lc.SetParent(s)
			s.Cases = append(s.Cases, lc)
			lc.Condition = expression(lc, c.Cond)
			createFunction(lc, c.Stmts)
			if len(lc.Statements) > 0 {
				_, ok := lc.Statements[len(lc.Statements)-1].(*lang.Break)
				if ok {
					lc.Statements = lc.Statements[:len(lc.Statements)-1]
				} else {
					f := &lang.Fallthrough{}
					f.SetParent(lc)
					lc.Statements = append(lc.Statements, f)
				}
			}

		case *stmt.Default:
			c := c.(*stmt.Default)
			d := &lang.Default{
				Vars:       make([]lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			d.SetParent(s)
			s.Cases = append(s.Cases, d)
			createFunction(d, c.Stmts)

			if len(d.Statements) > 0 {
				_, ok := d.Statements[len(d.Statements)-1].(*lang.Break)
				if ok {
					d.Statements = d.Statements[:len(d.Statements)-1]
				} else {
					f := &lang.Fallthrough{}
					f.SetParent(d)
					d.Statements = append(d.Statements, f)
				}
			}
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
		v := b.HasVariable(name)
		if v == nil {
			panic("Using undefined variable \"" + name + "\".")
		}
		return &lang.Variable{
			// Type will be taken from the right side.
			Type:      v.GetType(),
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
		v, isVar := expression(b, n.(*expr.PostInc).Variable).(*lang.Variable)
		if !isVar {
			panic(`Sadly enough, "++" requires variable, for now`)
		}
		v = b.HasVariable(v.Name)
		i := &lang.Inc{
			Var: v,
		}
		i.SetParent(b)
		return i

	case *expr.PostDec:
		i := &lang.Dec{
			Var: expression(b, n.(*expr.PostDec).Variable).(*lang.Variable),
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

	case *binary.SmallerOrEqual:
		p := n.(*binary.SmallerOrEqual)
		op := lang.CreateBinaryOp("<=", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.GreaterOrEqual:
		p := n.(*binary.GreaterOrEqual)
		op := lang.CreateBinaryOp(">=", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.Greater:
		p := n.(*binary.Greater)
		op := lang.CreateBinaryOp(">", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *binary.Identical:
		p := n.(*binary.Identical)
		op := lang.CreateBinaryOp("==", expression(b, p.Left), expression(b, p.Right))
		op.SetParent(b)
		return op

	case *expr.ConstFetch:
		cf := n.(*expr.ConstFetch)
		c := &lang.Const{
			Value: constructName(cf.Constant.(*name.Name)),
		}
		c.SetParent(b)
		return c

	// TODO: Add std functions to this parser, so it does not have to be
	// hacked like this.
	case *binary.Concat:
		c := n.(*binary.Concat)
		f := &lang.FunctionCall{
			Name:   "std.Concat",
			Args:   make([]lang.Expression, 0),
			Return: lang.String,
		}
		f.AddArg(expression(b, c.Left))
		f.AddArg(expression(b, c.Right))
		return f

	case *expr.FunctionCall:
		fc := n.(*expr.FunctionCall)

		n := constructName(fc.Function.(*name.Name))
		lf := gc.Get(n)
		if gc == nil {
			panic(n + " is not defined")
		}

		f := &lang.FunctionCall{
			Name:   n,
			Args:   make([]lang.Expression, 0),
			Return: gc.Get(n).GetType(),
		}

		al := fc.ArgumentList
		err := checkArguments(lf.Args, al.Arguments)
		if err != nil {
			panic(err)
		}
		for i, a := range al.Arguments {
			var arg lang.Expression
			if lf.Args[i].Reference {
				v := expression(b, a.(*node.Argument).Expr).(*lang.Variable)
				v.Reference = true
				arg = v
			} else {
				// TODO: Do not ignore information in Argument,
				// it has interesting information like if it is
				// send by reference and others.
				arg = expression(b, a.(*node.Argument).Expr)
			}
			f.AddArg(arg)
		}
		return f

	default:
		panic(`Something else uncatched.`)
	}
}

func checkArguments(vars []lang.Variable, call []node.Node) error {
	if len(vars) != len(call) {
		return errors.New("wrong argument count")
	}
	for i := 0; i < len(vars); i++ {
		_, isVar := call[i].(*node.Argument).Expr.(*expr.Variable)
		// This is something even PHP linter is aware of.
		if vars[i].Reference && !isVar {
			return errors.New("only variable can be parsed by reference")
		}
	}

	return nil
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
	switch res {
	case "func":
		res = "function"
	}
	return res
}
