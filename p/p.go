package p

import (
	"errors"
	"strings"

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
	// IdentifierName method is for this.
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
			n.SetParent(b)
			b.AddStatement(n)

		case *stmt.StmtList:
			list := lang.NewCode(b)
			b.AddStatement(list)
			createFunction(list, s.(*stmt.StmtList).Stmts)

		case *stmt.Expression:
			ex := makeExpression(b, s.(*stmt.Expression))
			b.AddStatement(ex)

		case *stmt.For:
			f := s.(*stmt.For)
			lf := lang.ConstructFor(b)

			if f.Init != nil {
				n := f.Init[0]
				ex := simpleExpression(lf, n)
				lf.Init = ex
			}

			if f.Cond != nil {
				n := f.Cond[0]
				err := lf.SetCond(expression(lf, n))
				if err != nil {
					panic(err)
				}
			}

			if f.Loop != nil {
				n := f.Loop[0]
				ex := simpleExpression(lf, n)
				lf.Loop = ex
			}

			createFunction(lf.Block, nodeList(f.Stmt))
			b.AddStatement(lf)

		case *stmt.While:
			w := s.(*stmt.While)
			lf := lang.ConstructFor(b)

			err := lf.SetCond(expression(lf, w.Cond))
			if err != nil {
				panic(err)
			}

			createFunction(lf.Block, nodeList(w.Stmt))
			b.AddStatement(lf)

		case *stmt.Do:
			w := s.(*stmt.Do)
			lf := lang.ConstructFor(b)

			createFunction(lf.Block, nodeList(w.Stmt))

			i := &lang.If{
				Vars: make([]lang.Variable, 0),
			}
			i.True = lang.NewCode(i)
			i.SetParent(lf)
			c := expression(i, w.Cond)
			neg := &lang.Negation{
				Right: c,
			}
			c.SetParent(neg)
			_ = i.SetCond(neg)
			neg.SetParent(i)
			i.True.AddStatement(&lang.Break{})
			lf.Block.AddStatement(i)

			b.AddStatement(lf)

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
				Expression: expression(b, s.(*stmt.Return).Expr),
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
				f.AddArg(expression(b, e))
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

func nodeList(n node.Node) []node.Node {
	list, ok := n.(*stmt.StmtList)
	if ok {
		return list.Stmts
	} else {
		return []node.Node{n}
	}
}

func constructIf(b lang.Block, i *stmt.If) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	err := nif.SetCond(expression(nif, i.Cond))
	if err != nil {
		panic(err)
	}

	nif.True = lang.NewCode(nif)
	createFunction(nif.True, nodeList(i.Stmt))

	lif := nif
	for _, ei := range i.ElseIf {
		lif.False = constructElif(b, ei.(*stmt.ElseIf))
		lif = lif.False.(*lang.If)
	}

	if i.Else == nil {
		return nif
	}

	e := i.Else.(*stmt.Else).Stmt
	switch e.(type) {
	case *stmt.If:
		lif.False = constructIf(lif, e.(*stmt.If))

	default:
		c := lang.NewCode(lif)
		createFunction(c, nodeList(e))
		lif.False = c
	}
	return nif
}

func constructElif(b lang.Block, i *stmt.ElseIf) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	err := nif.SetCond(expression(nif, i.Cond))
	if err != nil {
		panic(err)
	}

	nif.True = lang.NewCode(nif)
	createFunction(nif.True, nodeList(i.Stmt))
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

func makeExpression(b lang.Block, e *stmt.Expression) lang.Node {
	s := statement(b, e.Expr)
	if s != nil {
		return s
	}
	return complexExpression(b, e.Expr)
}

func simpleExpression(b lang.Block, n node.Node) lang.Node {
	s := statement(b, n)
	if s != nil {
		return s
	}

	e := expression(b, n)
	if e != nil {
		return e
	}

	switch n.(type) {
	case *assign.Assign:
		a := n.(*assign.Assign)

		r := expression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		n := identifierName(a.Variable.(*expr.Variable))
		return buildAssignment(b, n, r)
	}
	panic(`SimpleExpression: something else uncatched.`)
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

		r := complexExpression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		la, ok := r.(*lang.Assign)
		if ok {
			b.AddStatement(la)
			r = la.Left()
		}

		n := identifierName(a.Variable.(*expr.Variable))
		return buildAssignment(b, n, r)
	}
	panic(`ComplexExpression: something else uncatched.`)
}

func statement(b lang.Block, n node.Node) lang.Node {
	switch n.(type) {
	case *expr.PostInc:
		v, isVar := expression(b, n.(*expr.PostInc).Variable).(*lang.Variable)
		if !isVar {
			panic(`"++" requires variable.`)
		}
		v = b.HasVariable(v.Name)
		i := &lang.Inc{
			Var: v,
		}
		i.SetParent(b)
		return i

	case *expr.PostDec:
		v, ok := expression(b, n.(*expr.PostDec).Variable).(*lang.Variable)
		if !ok {
			panic(`"--" requires variable.`)
		}

		i := &lang.Dec{
			Var: v,
		}
		i.SetParent(b)
		return i
	}
	return nil
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
			Type:      v.GetType(),
			Name:      name,
			Const:     false,
			Reference: false,
		}

	case *scalar.Encapsed:
		e := n.(*scalar.Encapsed)
		f := &lang.FunctionCall{
			Name:   "fmt.Sprintf",
			Args:   make([]lang.Expression, 1),
			Return: lang.String,
		}
		s := &lang.Str{
			Value: "\"",
		}
		s.SetParent(f)
		for _, p := range e.Parts {
			switch p.(type) {
			case *scalar.EncapsedStringPart:
				s.Value += p.(*scalar.EncapsedStringPart).Value

			case *expr.Variable:
				vn := identifierName(p.(*expr.Variable))
				v := b.HasVariable(vn)
				if v == nil || v.GetType() == lang.Void {
					panic(vn + " is not defined.")
				}
				// TODO: Type could know this.
				switch v.GetType() {
				case lang.Int:
					s.Value += "%d"

				case lang.Float64:
					s.Value += "%g"

				case lang.String:
					s.Value += "%s"
				}
				f.AddArg(v)

			case *expr.ArrayDimFetch:
				adf := p.(*expr.ArrayDimFetch)
				vn := identifierName(adf.Variable.(*expr.Variable))
				v := b.HasVariable(vn)
				if v == nil || v.GetType() == lang.Void {
					panic(vn + " is not defined.")
				}

				fa := &lang.FetchArr{
					Arr:   v,
					Index: expression(b, adf.Dim),
				}
				fa.Arr.SetParent(fa)
				fa.Index.SetParent(fa)
				fa.SetParent(fa)

				// TODO: Type could know this.
				switch fa.GetType() {
				case lang.Int:
					s.Value += "%d"

				case lang.Float64:
					s.Value += "%g"

				case lang.String:
					s.Value += "%s"
				}
				f.AddArg(fa)
			}
		}
		s.Value += "\""
		f.Args[0] = s
		f.SetParent(b)
		return f

	case *expr.UnaryPlus:
		e := expression(b, n.(*expr.UnaryPlus).Expr)
		e.SetParent(b)
		return e

	case *expr.UnaryMinus:
		m := &lang.UnaryMinus{
			Expr: expression(b, n.(*expr.UnaryMinus).Expr),
		}
		m.SetParent(b)
		return m

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
		if s[0] == '\'' && s[len(s)-1] == '\'' {
			s = strings.ReplaceAll(s, "\\", "\\\\")
			s = strings.ReplaceAll(s, "\"", "\\\"")
			s = strings.ReplaceAll(s, "'", "\"")
		}
		str := &lang.Str{
			Value: s,
		}
		str.SetParent(b)
		return str

	case *expr.ShortArray:
		a := n.(*expr.ShortArray)
		items := make([]lang.Expression, 0)
		for _, i := range a.Items {
			items = append(items, expression(b, i.(*expr.ArrayItem).Val))
		}
		if len(items) == 0 {
			panic(`Cannot decide type, empty array.`)
		}
		typ := items[0].GetType()
		for _, i := range items {
			if i.GetType() != typ {
				panic(`Type is not the same for every element of the array.`)
			}
		}

		al := &lang.Array{
			Values: items,
			Type:   typ,
		}

		al.SetParent(b)
		return al

	case *expr.ArrayDimFetch:
		adf := n.(*expr.ArrayDimFetch)
		v, ok := expression(b, adf.Variable).(*lang.Variable)
		if !ok {
			panic(`Expected variable to be indexed.`)
		}
		fa := &lang.FetchArr{
			Arr:   v,
			Index: expression(b, adf.Dim),
		}
		fa.Arr.SetParent(fa)
		fa.Index.SetParent(fa)
		fa.SetParent(fa)
		return fa

	case *binary.Plus:
		p := n.(*binary.Plus)
		op, err := lang.CreateBinaryOp("+", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.Minus:
		p := n.(*binary.Minus)
		op, err := lang.CreateBinaryOp("-", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.Mul:
		p := n.(*binary.Mul)
		op, err := lang.CreateBinaryOp("*", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.Smaller:
		p := n.(*binary.Smaller)
		op, err := lang.CreateBinaryOp("<", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.SmallerOrEqual:
		p := n.(*binary.SmallerOrEqual)
		op, err := lang.CreateBinaryOp("<=", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.GreaterOrEqual:
		p := n.(*binary.GreaterOrEqual)
		op, err := lang.CreateBinaryOp(">=", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.Greater:
		p := n.(*binary.Greater)
		op, err := lang.CreateBinaryOp(">", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
		op.SetParent(b)
		return op

	case *binary.Identical:
		p := n.(*binary.Identical)
		op, err := lang.CreateBinaryOp("==", expression(b, p.Left), expression(b, p.Right))
		if err != nil {
			panic(err)
		}
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
		al := fc.ArgumentList

		n := constructName(fc.Function.(*name.Name))

		if n == "array_push" {
			if len(al.Arguments) < 2 {
				panic(`array_push requires atlast two arguments`)
			}
			v, ok := expression(b, al.Arguments[0].(*node.Argument).Expr).(*lang.Variable)
			if !ok {
				panic(`First argument has to be a variable.`)
			}
			vars := []lang.Expression{}
			for _, v := range al.Arguments[1:] {
				vars = append(vars, expression(b, v.(*node.Argument).Expr))
			}

			a := createArrayPush(b, v, vars)
			a.SetParent(b)
			return a
		}

		lf := gc.Get(n)
		if lf == nil {
			panic(n + " is not defined")
		}

		f := &lang.FunctionCall{
			Name:   n,
			Args:   make([]lang.Expression, 0),
			Return: gc.Get(n).GetType(),
		}

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
	}
	return nil
}

func checkArguments(vars []lang.Variable, call []node.Node) error {
	if len(vars) != len(call) {
		return errors.New("wrong argument count")
	}
	for i := 0; i < len(vars); i++ {
		_, isVar := call[i].(*node.Argument).Expr.(*expr.Variable)
		// This is something even PHP linter is aware of.
		if vars[i].Reference && !isVar {
			return errors.New("only variable can be passed by reference")
		}
	}

	return nil
}

func createArrayPush(b lang.Block, v *lang.Variable, vals []lang.Expression) *lang.Assign {
	if b.HasVariable(v.Name) == nil {
		panic(v.Name + " is not defined.")
	}

	f := &lang.FunctionCall{
		Name:   "append",
		Args:   []lang.Expression{v},
		Return: v.GetType(),
	}
	for _, val := range vals {
		if val.GetType() != f.GetType() {
			panic(f.GetType() + " expected, " + val.GetType() + " found.")
		}
		val.SetParent(f)
		f.Args = append(f.Args, val)
	}

	return lang.CreateAssign(v, f)
}

func buildAssignment(parent lang.Block, varName string, right lang.Expression) *lang.Assign {
	t := right.GetType()
	if t == lang.Void {
		panic("Cannot assign \"void\" " + "to \"" + varName + "\".")
	}
	av := &lang.Variable{
		Type:      t,
		Name:      varName,
		Const:     false,
		Reference: false,
	}

	fd := true
	if v := parent.HasVariable(varName); v != nil {
		fd = false
		if v.GetType() != t {
			if parent.DefinesVariable(varName) == nil {
				fd = true
				parent.DefineVariable(*av)
			} else {
				panic("Invalid assignment, \"" + v.GetType() + "\" expected, \"" + t + "\" given.")
			}
		}
	} else {
		parent.DefineVariable(*av)
	}
	as := lang.CreateAssign(av, right)
	as.FirstDefinition = fd
	as.SetParent(parent)

	return as
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
