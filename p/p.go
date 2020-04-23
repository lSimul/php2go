package p

import (
	"errors"
	"fmt"
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
var translator, functionTranslator NameTranslation
var useGlobalContext = false

func Run(r *node.Root) *lang.GlobalContext {
	gc = lang.NewGlobalContext()
	translator = NewNameTranslator()
	functionTranslator = NewFunctionTranslator()
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
	useGlobalContext = true
	createFunction(&main.Body, ms)
	useGlobalContext = false
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
	if fc == nil {
		return nil
	}

	// TODO: IdentifierName method is for this. (is it still relevant?)
	// TODO: Make sure visibility is set as it should be.
	n := functionTranslator.Translate(fc.FunctionName.(*node.Identifier).Value, Private)
	f := lang.NewFunc(n)
	f.SetParent(gc)

	for _, pr := range fc.Params {
		p := pr.(*node.Parameter)
		v := newVariable(
			identifierName(p.Variable.(*expr.Variable)),
			constructName(p.VariableType.(*name.Name), true),
			false)
		f.Args = append(f.Args, v)
	}

	if fc.ReturnType != nil {
		n := constructName(fc.ReturnType.(*name.Name), true)
		if n == "void" {
			n = lang.Void
		}
		f.Return = n
	}

	return f
}

func mainDef() *lang.Function {
	f := lang.NewFunc("main")
	f.SetParent(gc)
	return f
}

func createFunction(b lang.Block, stmts []node.Node) {
	for _, s := range stmts {
		switch s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			f := &lang.FunctionCall{
				Name: "fmt.Print",
				Args: make([]lang.Expression, 0),
			}
			s := &lang.Str{
				Value: fmt.Sprintf("`%s`", s.(*stmt.InlineHtml).Value),
			}
			s.SetParent(f)
			f.AddArg(s)
			f.SetParent(b)
			b.AddStatement(f)

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
				e := expression(lf, n)
				if e.GetType() != lang.Bool {
					e = &lang.FunctionCall{
						Name:   "std.Truthy",
						Args:   []lang.Expression{e},
						Return: lang.Bool,
					}
					e.SetParent(lf)
				}
				err := lf.SetCond(e)
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

			e := expression(lf, w.Cond)
			if e.GetType() != lang.Bool {
				e = &lang.FunctionCall{
					Name:   "std.Truthy",
					Args:   []lang.Expression{e},
					Return: lang.Bool,
				}
				e.SetParent(lf)
			}

			err := lf.SetCond(e)
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
				Vars: make([]*lang.Variable, 0),
			}
			i.True = lang.NewCode(i)
			i.SetParent(lf)
			// TODO: Negation should work only with
			// boolean values.
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

		case *stmt.Foreach:
			f := s.(*stmt.Foreach)
			lf := &lang.Foreach{}
			lf.SetParent(b)
			lf.Iterated = expression(lf, f.Expr)
			if f.Key != nil {
				name := identifierName(f.Key.(*expr.Variable))
				lf.Key = newVariable(name, lang.Int, false)
			}

			name := identifierName(f.Variable.(*expr.Variable))
			lf.Value = *newVariable(name, lf.Iterated.GetType(), false)
			lf.Block = lang.NewCode(lf)
			createFunction(lf.Block, nodeList(f.Stmt))

			lf.SetParent(b)
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
		if n == nil {
			return []node.Node{}
		} else {
			return []node.Node{n}
		}
	}
}

func constructIf(b lang.Block, i *stmt.If) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	expr := expression(nif, i.Cond)
	if expr.GetType() != lang.Bool {
		expr = &lang.FunctionCall{
			Name:   "std.Truthy",
			Args:   []lang.Expression{expr},
			Return: lang.Bool,
		}
		expr.SetParent(nif)
	}
	err := nif.SetCond(expr)
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
	e := expression(nif, i.Cond)
	if e.GetType() != lang.Bool {
		e = &lang.FunctionCall{
			Name:   "std.Truthy",
			Args:   []lang.Expression{e},
			Return: lang.Bool,
		}
		e.SetParent(nif)
	}
	err := nif.SetCond(e)
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
				Vars:       make([]*lang.Variable, 0),
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
				Vars:       make([]*lang.Variable, 0),
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
			r = lang.NewVarRef(la.Left(), la.GetType())
		}

		n := identifierName(a.Variable.(*expr.Variable))
		return buildAssignment(b, n, r)
	}
	panic(`ComplexExpression: something else uncatched.`)
}

func statement(b lang.Block, n node.Node) lang.Node {
	var v *lang.VarRef
	var ok bool

	inc := func() lang.Node {
		if !ok {
			panic(`"++" requires variable.`)
		}
		if b.HasVariable(v.V.Name) == nil {
			panic(fmt.Sprintf("'%s' is not defined.", v.V.Name))
		}
		return lang.NewInc(b, v)
	}
	dec := func() lang.Node {
		if !ok {
			panic(`"--" requires variable.`)
		}
		if b.HasVariable(v.V.Name) == nil {
			panic(fmt.Sprintf("'%s' is not defined.", v.V.Name))
		}
		return lang.NewDec(b, v)
	}

	switch n.(type) {
	case *expr.PreInc:
		v, ok = expression(b, n.(*expr.PreInc).Variable).(*lang.VarRef)
		return inc()
	case *expr.PostInc:
		v, ok = expression(b, n.(*expr.PostInc).Variable).(*lang.VarRef)
		return inc()

	case *expr.PreDec:
		v, ok = expression(b, n.(*expr.PreDec).Variable).(*lang.VarRef)
		return dec()
	case *expr.PostDec:
		v, ok = expression(b, n.(*expr.PostDec).Variable).(*lang.VarRef)
		return dec()
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
		return lang.NewVarRef(v, v.GetType())

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
				f.AddArg(lang.NewVarRef(v, v.GetType()))

			case *expr.ArrayDimFetch:
				adf := p.(*expr.ArrayDimFetch)
				vn := identifierName(adf.Variable.(*expr.Variable))
				v := b.HasVariable(vn)
				if v == nil || v.GetType() == lang.Void {
					panic(vn + " is not defined.")
				}

				scalar := &lang.FunctionCall{
					Name:   "array.NewScalar",
					Args:   []lang.Expression{expression(b, adf.Dim)},
					Return: lang.String,
				}
				fc := &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.At", v),
					Args:   []lang.Expression{scalar},
					Return: ArrayItem(v.GetType()),
				}
				fmt.Println(fc.Return)
				scalar.SetParent(fc)
				fc.SetParent(b)

				// TODO: Type could know this.
				switch fc.GetType() {
				case lang.Int:
					s.Value += "%d"

				case lang.Float64:
					s.Value += "%g"

				case lang.String:
					s.Value += "%s"
				}
				f.AddArg(fc)
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
		m.Expr.SetParent(m)
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
			v := i.(*expr.ArrayItem).Val
			if v == nil {
				continue
			}
			items = append(items, expression(b, v))
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

		fc := &lang.FunctionCall{
			Name:   "array.New" + FirstUpper(typ),
			Args:   items,
			Return: ArrayType(typ),
		}

		fc.SetParent(b)
		return fc

	case *expr.ArrayDimFetch:
		adf := n.(*expr.ArrayDimFetch)
		v, ok := expression(b, adf.Variable).(*lang.VarRef)
		if !ok {
			panic(`Expected variable to be indexed.`)
		}

		scalar := &lang.FunctionCall{
			Name:   "array.NewScalar",
			Args:   []lang.Expression{expression(b, adf.Dim)},
			Return: lang.String,
		}
		fc := &lang.FunctionCall{
			Name:   fmt.Sprintf("%s.At", v),
			Args:   []lang.Expression{scalar},
			Return: ArrayItem(v.GetType()),
		}
		scalar.SetParent(fc)
		fc.SetParent(b)
		return fc

	case *binary.Plus:
		p := n.(*binary.Plus)
		return binaryOp(b, "+", p.Left, p.Right)

	case *binary.Minus:
		p := n.(*binary.Minus)
		return binaryOp(b, "-", p.Left, p.Right)

	case *binary.Mul:
		p := n.(*binary.Mul)
		return binaryOp(b, "*", p.Left, p.Right)

	case *binary.Smaller:
		p := n.(*binary.Smaller)
		return binaryOp(b, "<", p.Left, p.Right)

	case *binary.SmallerOrEqual:
		p := n.(*binary.SmallerOrEqual)
		return binaryOp(b, "<=", p.Left, p.Right)

	case *binary.GreaterOrEqual:
		p := n.(*binary.GreaterOrEqual)
		return binaryOp(b, ">=", p.Left, p.Right)

	case *binary.Greater:
		p := n.(*binary.Greater)
		return binaryOp(b, ">", p.Left, p.Right)

	case *binary.Identical:
		p := n.(*binary.Identical)
		return binaryOp(b, "==", p.Left, p.Right)

	case *expr.ConstFetch:
		cf := n.(*expr.ConstFetch)
		c := &lang.Const{
			Value: constructName(cf.Constant.(*name.Name), true),
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

		n := constructName(fc.Function.(*name.Name), false)
		if ok := PHPFunctions[n]; ok {
			if n == "array_push" {
				if len(al.Arguments) < 2 {
					panic(`array_push requires atlast two arguments`)
				}
				v, ok := expression(b, al.Arguments[0].(*node.Argument).Expr).(*lang.VarRef)
				if !ok {
					panic(`First argument has to be a variable.`)
				}
				typ := ArrayItem(v.GetType())

				vars := []lang.Expression{}
				for _, arg := range al.Arguments[1:] {
					v := expression(b, arg.(*node.Argument).Expr)
					if v.GetType() != typ {
						panic(`Cannot push this type.`)
					}
					vars = append(vars, v)
				}

				fc := &lang.FunctionCall{
					Name:   v.V.Name + ".Push",
					Args:   vars,
					Return: lang.Int,
				}

				fc.SetParent(b)
				return fc
			}
		}

		// TODO: Remove this ugly temporary solution, translating has to be smarter.
		n = constructName(fc.Function.(*name.Name), false)
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
		// TODO: Check for passing by reference.
		for _, a := range al.Arguments {
			// TODO: Do not ignore information in Argument,
			// it has interesting information like if it is
			// send by reference and others.
			arg := expression(b, a.(*node.Argument).Expr)
			f.AddArg(arg)
		}
		return f
	}
	return nil
}

func binaryOp(b lang.Block, op string, left, right node.Node) lang.Expression {
	res, err := lang.NewBinaryOp(op, expression(b, left), expression(b, right))
	if err != nil {
		panic(err)
	}
	res.SetParent(b)
	return res
}

func checkArguments(vars []*lang.Variable, call []node.Node) error {
	if len(vars) != len(call) {
		return errors.New("wrong argument count")
	}
	// TODO: Check if arguments are passed by reference, make sure
	// that is done only with variables.
	return nil
}

func buildAssignment(parent lang.Block, name string, right lang.Expression) *lang.Assign {
	t := right.GetType()
	if t == lang.Void {
		panic("Cannot assign \"void\" " + "to \"" + name + "\".")
	}

	if useGlobalContext {
		name = translator.Translate(name, Public)
	} else {
		name = translator.Translate(name, Private)
	}

	v := parent.HasVariable(name)
	fd := false
	if v == nil {
		v = newVariable(name, t, false)
		if useGlobalContext {
			gc.DefineVariable(v)
		} else {
			parent.DefineVariable(v)
		}
		fd = true
	} else if v.CurrentType != t {
		if useGlobalContext {
			// TODO: This is something I solve in the define variable.
			v.Type = lang.Anything
			v.CurrentType = t
		} else {
			if v.FirstDefinition.Parent() == parent {
				v.CurrentType = t
			} else {
				v = newVariable(name, t, false)
				fd = true
			}
		}
		parent.DefineVariable(v)
	}

	as, err := lang.NewAssign(v, right)
	if err != nil {
		panic(err)
	}

	if useGlobalContext && fd {
		v.FirstDefinition = gc
		fd = false
	}
	if fd {
		v.FirstDefinition = as
	}

	as.FirstDefinition = fd
	as.SetParent(parent)

	return as
}

func newVariable(name, typ string, isConst bool) *lang.Variable {
	// TODO: This fixed "Public" does not look right.
	// It should probably require already unique name.
	name = translator.Translate(name, Public)
	return &lang.Variable{
		Name:  name,
		Type:  typ,
		Const: isConst,

		CurrentType: typ,
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
		n := v.VarName.(*node.Identifier).Value
		if useGlobalContext {
			return translator.Translate(n, Public)
		} else {
			return translator.Translate(n, Private)
		}

	default:
		panic(`Variable name is not defined as a simple string.`)
	}
}

func constructName(nm *name.Name, translate bool) string {
	s := ""
	for _, n := range nm.Parts {
		s += n.(*name.NamePart).Value
	}
	if !translate {
		return s
	}
	if useGlobalContext {
		return functionTranslator.Translate(s, Public)
	}
	return functionTranslator.Translate(s, Private)
}
