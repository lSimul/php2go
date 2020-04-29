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

type parser struct {
	translator         NameTranslation
	functionTranslator NameTranslation

	gc               *lang.GlobalContext
	useGlobalContext bool
}

func NewParser(v, f NameTranslation) *parser {
	return &parser{
		translator:         v,
		functionTranslator: f,
	}
}

func (p *parser) Run(r *node.Root) *lang.GlobalContext {
	p.gc = lang.NewGlobalContext()
	ms, fs := sanitizeRootStmts(r)

	for _, s := range fs {
		f := p.funcDef(&s)
		p.gc.Add(f)
	}
	for _, s := range fs {
		f := p.funcDef(&s)
		p.createFunction(&f.Body, s.Stmts)
		p.gc.Add(f)
	}

	main := p.mainDef()
	p.gc.Add(main)
	p.useGlobalContext = true
	p.createFunction(&main.Body, ms)
	p.useGlobalContext = false
	return p.gc
}

// SanitizeRootStmts splits statements based on their type,
// functions go to the one array, rest to the other one.
// This makes function root more straight forward, main will
// not be longer split up by functions.
func sanitizeRootStmts(r *node.Root) ([]node.Node, []stmt.Function) {
	main := make([]node.Node, 0)
	functions := make([]stmt.Function, 0)

	for _, s := range r.Stmts {
		switch s := s.(type) {
		case *stmt.Function:
			functions = append(functions, *s)
		default:
			main = append(main, s)
		}
	}

	return main, functions
}

func (parser *parser) funcDef(fc *stmt.Function) *lang.Function {
	if fc == nil {
		return nil
	}

	// TODO: IdentifierName method is for this. (is it still relevant?)
	// TODO: Make sure visibility is set as it should be.
	n := parser.functionTranslator.Translate(fc.FunctionName.(*node.Identifier).Value, Private)
	f := lang.NewFunc(n)
	f.SetParent(parser.gc)

	for _, pr := range fc.Params {
		p := pr.(*node.Parameter)
		v := parser.newVariable(
			parser.identifierName(p.Variable.(*expr.Variable)),
			parser.constructName(p.VariableType.(*name.Name), true),
			false)
		f.Args = append(f.Args, v)
	}

	if fc.ReturnType != nil {
		n := parser.constructName(fc.ReturnType.(*name.Name), true)
		if n == "void" {
			n = lang.Void
		}
		f.Return = n
	}

	return f
}

func (p *parser) mainDef() *lang.Function {
	f := lang.NewFunc("main")
	f.SetParent(p.gc)
	return f
}

func (parser *parser) createFunction(b lang.Block, stmts []node.Node) {
	for _, s := range stmts {
		switch s := s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			f := &lang.FunctionCall{
				Name: "fmt.Print",
				Args: make([]lang.Expression, 0),
			}

			parser.gc.RequireNamespace("fmt")

			str := &lang.Str{
				Value: fmt.Sprintf("`%s`", s.Value),
			}
			str.SetParent(f)
			f.AddArg(str)
			f.SetParent(b)
			b.AddStatement(f)

		case *stmt.StmtList:
			list := lang.NewCode(b)
			b.AddStatement(list)
			parser.createFunction(list, s.Stmts)

		case *stmt.Expression:
			ex := parser.makeExpression(b, s)
			b.AddStatement(ex)

		case *stmt.For:
			lf := lang.ConstructFor(b)

			if s.Init != nil {
				n := s.Init[0]
				ex := parser.simpleExpression(lf, n)
				lf.Init = ex
			}

			if s.Cond != nil {
				c := parser.conditionExpr(lf, s.Cond[0])
				err := lf.SetCond(c)
				if err != nil {
					panic(err)
				}
			}

			if s.Loop != nil {
				n := s.Loop[0]
				ex := parser.simpleExpression(lf, n)
				lf.Loop = ex
			}

			parser.createFunction(lf.Block, nodeList(s.Stmt))
			b.AddStatement(lf)

		case *stmt.While:
			lf := lang.ConstructFor(b)
			c := parser.conditionExpr(lf, s.Cond)

			err := lf.SetCond(c)
			if err != nil {
				panic(err)
			}

			parser.createFunction(lf.Block, nodeList(s.Stmt))
			b.AddStatement(lf)

		case *stmt.Do:
			lf := lang.ConstructFor(b)

			parser.createFunction(lf.Block, nodeList(s.Stmt))

			i := &lang.If{
				Vars: make([]*lang.Variable, 0),
			}
			i.True = lang.NewCode(i)
			i.SetParent(lf)
			// TODO: Negation should work only with
			// boolean values.
			c := parser.expression(i, s.Cond)
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
			lf := &lang.Foreach{}
			lf.SetParent(b)
			lf.Block = lang.NewCode(lf)

			iterated := parser.expression(lf, s.Expr)
			if !IsArray(iterated.Type()) {
				panic(`Only arrays can be iterated.`)
			}

			var fnName string
			switch i := iterated.(type) {
			case *lang.VarRef:
				fnName = i.V.Name
			case *lang.FunctionCall:
				fnName = i.String()

			default:
				panic(`Uncatched type of iterated.`)
			}

			var it *lang.FunctionCall
			// Easy chain, without array.Pair
			if s.Key == nil {
				it = &lang.FunctionCall{
					Name: fnName + ".Iter",
					// TODO: Set up return type.
				}
				name := parser.identifierName(s.Variable.(*expr.Variable))
				lf.Value = *parser.newVariable(name, ArrayItem(iterated.Type()), false)
			} else {
				name := parser.identifierName(s.Key.(*expr.Variable))
				// TODO: Define type scalar which is being
				// formated as string.
				k := parser.newVariable(name, lang.String, false)
				n := parser.identifierName(s.Variable.(*expr.Variable))
				v := parser.newVariable(n, ArrayItem(iterated.Type()), false)

				it = &lang.FunctionCall{
					Name: fnName + ".KeyIter",
					// TODO: Set up return type.
				}

				// TODO: This is not lang.Void
				lf.Value = *parser.newVariable("pair", lang.Void, false)

				// TODO: I do not have this part of code under control.
				// Accessing struct elements is out of my reach right now.
				if k != nil {
					pairK := parser.newVariable(lf.Value.Name+".K", lang.String, true)
					s, err := lang.NewAssign(k, lang.NewVarRef(pairK, pairK.Type()))
					if err != nil {
						panic(err)
					}
					s.FirstDefinition = true
					lf.Block.AddStatement(s)
					lf.Block.DefineVariable(k)
				}

				pairV := parser.newVariable(lf.Value.Name+".V", ArrayItem(iterated.Type()), true)
				s, err := lang.NewAssign(v, lang.NewVarRef(pairV, pairV.Type()))
				if err != nil {
					panic(err)
				}
				s.FirstDefinition = true
				lf.Block.AddStatement(s)
				lf.Block.DefineVariable(v)
			}
			it.SetParent(lf)
			lf.Iterated = it

			parser.createFunction(lf.Block, nodeList(s.Stmt))

			lf.SetParent(b)
			b.AddStatement(lf)

		case *stmt.If:
			i := parser.constructIf(b, s)
			b.AddStatement(i)

		case *stmt.Switch:
			sw := &lang.Switch{
				Cases: make([]lang.Node, 0),
			}
			sw.SetParent(b)
			sw.Condition = parser.expression(sw, s.Cond)
			parser.constructSwitch(sw, s.CaseList)
			b.AddStatement(sw)

		case *stmt.Return:
			r := &lang.Return{
				Expression: parser.expression(b, s.Expr),
			}
			b.AddStatement(r)

		case *stmt.Echo:
			f := &lang.FunctionCall{
				Name: "fmt.Print",
				Args: make([]lang.Expression, 0),
			}

			parser.gc.RequireNamespace("fmt")

			for _, e := range s.Exprs {
				// TODO: Do not ignore information in Argument,
				// it has interesting information like if it is
				// send by reference and others.
				f.AddArg(parser.expression(b, e))
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
			// parser.statemtn contains also "my" statements
			// like {inc,dec}rements, so the structure is not
			// 1:1 with the stuff which come from the parser.
			n := parser.statement(b, s)
			if n == nil {
				panic(`Unexpected statement`)
			}
			n.SetParent(b)
			b.AddStatement(n)
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

func (p *parser) constructIf(b lang.Block, i *stmt.If) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	c := p.conditionExpr(nif, i.Cond)

	err := nif.SetCond(c)
	if err != nil {
		panic(err)
	}

	nif.True = lang.NewCode(nif)
	p.createFunction(nif.True, nodeList(i.Stmt))

	lif := nif
	for _, ei := range i.ElseIf {
		lif.False = p.constructElif(b, ei.(*stmt.ElseIf))
		lif = lif.False.(*lang.If)
	}

	if i.Else == nil {
		return nif
	}

	e := i.Else.(*stmt.Else).Stmt
	switch t := e.(type) {
	case *stmt.If:
		lif.False = p.constructIf(lif, t)

	default:
		c := lang.NewCode(lif)
		p.createFunction(c, nodeList(t))
		lif.False = c
	}
	return nif
}

func (p *parser) constructElif(b lang.Block, i *stmt.ElseIf) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	c := p.conditionExpr(nif, i.Cond)

	err := nif.SetCond(c)
	if err != nil {
		panic(err)
	}

	nif.True = lang.NewCode(nif)
	p.createFunction(nif.True, nodeList(i.Stmt))
	return nif
}

func (parser *parser) constructSwitch(s *lang.Switch, cl *stmt.CaseList) {
	for _, c := range cl.Cases {
		switch c := c.(type) {
		case *stmt.Case:
			lc := &lang.Case{
				Vars:       make([]*lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			lc.SetParent(s)
			s.Cases = append(s.Cases, lc)
			lc.Condition = parser.expression(lc, c.Cond)
			parser.createFunction(lc, c.Stmts)
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
			d := &lang.Default{
				Vars:       make([]*lang.Variable, 0),
				Statements: make([]lang.Node, 0),
			}
			d.SetParent(s)
			s.Cases = append(s.Cases, d)
			parser.createFunction(d, c.Stmts)

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

func (p *parser) makeExpression(b lang.Block, e *stmt.Expression) lang.Node {
	s := p.statement(b, e.Expr)
	if s != nil {
		return s
	}
	return p.complexExpression(b, e.Expr)
}

func (p *parser) simpleExpression(b lang.Block, n node.Node) lang.Node {
	s := p.statement(b, n)
	if s != nil {
		return s
	}

	e := p.expression(b, n)
	if e != nil {
		return e
	}

	if a, ok := n.(*assign.Assign); ok {
		r := p.expression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		n := p.identifierName(a.Variable.(*expr.Variable))
		return p.buildAssignment(b, n, r)
	}
	panic(`SimpleExpression: something else uncatched.`)
}

func (parser *parser) complexExpression(b lang.Block, n node.Node) lang.Expression {
	e := parser.expression(b, n)
	if e != nil {
		return e
	}

	a, ok := n.(*assign.Assign)
	if !ok {
		panic(`ComplexExpression: something else uncatched.`)
	}

	// Every expression should have return value.
	// Otherwise I cannot say what the assigned value will have.
	r := parser.complexExpression(b, a.Expression)
	if r == nil {
		panic(`Missing right side for assignment.`)
	}

	la, ok := r.(*lang.Assign)
	if ok {
		b.AddStatement(la)
		r = lang.NewVarRef(la.Left(), la.Type())
	}

	switch v := a.Variable.(type) {
	case (*expr.Variable):
		n := parser.identifierName(v)
		return parser.buildAssignment(b, n, r)

	case (*expr.ArrayDimFetch):
		vn := parser.identifierName(v.Variable.(*expr.Variable))
		vr := b.HasVariable(vn)
		if vr == nil || vr.Type() == lang.Void {
			panic(vn + " is not defined.")
		}

		if l, r := ArrayItem(vr.Type()), r.Type(); l != r {
			panic(fmt.Sprintf("Array editing: '%s' expected, '%s' given.", l, r))
		}

		var fc *lang.FunctionCall
		if v.Dim == nil {
			fc = &lang.FunctionCall{
				Name:   fmt.Sprintf("%s.Add", vr),
				Args:   []lang.Expression{r},
				Return: vr.Type(),
			}
			fc.SetParent(b)
		} else {
			scalar := &lang.FunctionCall{
				Name:   "array.NewScalar",
				Args:   []lang.Expression{parser.expression(b, v.Dim)},
				Return: lang.String,
			}
			fc = &lang.FunctionCall{
				Name:   fmt.Sprintf("%s.Edit", vr),
				Args:   []lang.Expression{scalar, r},
				Return: vr.Type(),
			}

			parser.gc.RequireNamespace("array")

			scalar.SetParent(fc)
			fc.SetParent(b)
		}
		return fc

	default:
		panic(fmt.Sprintf("Unexpected left side: %v", a))
	}
}

func (parser *parser) statement(b lang.Block, n node.Node) lang.Node {
	// Inpisration from parser.expression, *expr.ArrayDimFetch.
	switch s := n.(type) {
	case *stmt.Unset:
		first := true
		var ret lang.Node
		for _, v := range s.Vars {
			adf, ok := v.(*expr.ArrayDimFetch)
			if !ok {
				panic(`Only arrays are accepted for unset.`)
			}
			vn := parser.identifierName(adf.Variable.(*expr.Variable))
			v := b.HasVariable(vn)
			if v == nil || v.Type() == lang.Void {
				panic(vn + " is not defined.")
			}
			scalar := &lang.FunctionCall{
				Name:   "array.NewScalar",
				Args:   []lang.Expression{parser.expression(b, adf.Dim)},
				Return: lang.String,
			}

			parser.gc.RequireNamespace("array")

			fc := &lang.FunctionCall{
				Name: fmt.Sprintf("%s.Unset", v),
				Args: []lang.Expression{scalar},
			}
			if first {
				ret = fc
			} else {
				b.AddStatement(fc)
			}
		}
		return ret
	}

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

	switch e := n.(type) {
	case *expr.PreInc:
		v, ok = parser.expression(b, e.Variable).(*lang.VarRef)
		return inc()
	case *expr.PostInc:
		v, ok = parser.expression(b, e.Variable).(*lang.VarRef)
		return inc()

	case *expr.PreDec:
		v, ok = parser.expression(b, e.Variable).(*lang.VarRef)
		return dec()
	case *expr.PostDec:
		v, ok = parser.expression(b, e.Variable).(*lang.VarRef)
		return dec()
	}
	return nil
}

func (parser *parser) expression(b lang.Block, n node.Node) lang.Expression {
	switch e := n.(type) {
	case *expr.Variable:
		name := parser.identifierName(e)
		v := b.HasVariable(name)
		if v == nil {
			panic("Using undefined variable \"" + name + "\".")
		}
		return lang.NewVarRef(v, v.Type())

	case *scalar.Encapsed:
		f := &lang.FunctionCall{
			Name:   "fmt.Sprintf",
			Args:   make([]lang.Expression, 1),
			Return: lang.String,
		}

		parser.gc.RequireNamespace("fmt")

		s := &lang.Str{
			Value: "\"",
		}
		s.SetParent(f)
		for _, p := range e.Parts {
			switch p := p.(type) {
			case *scalar.EncapsedStringPart:
				s.Value += p.Value

			case *expr.Variable:
				vn := parser.identifierName(p)
				v := b.HasVariable(vn)
				if v == nil || v.Type() == lang.Void {
					panic(vn + " is not defined.")
				}
				// TODO: Type could know this.
				switch v.Type() {
				case lang.Int:
					s.Value += "%d"

				case lang.Float64:
					s.Value += "%g"

				case lang.String:
					s.Value += "%s"
				}
				f.AddArg(lang.NewVarRef(v, v.Type()))

			case *expr.ArrayDimFetch:
				vn := parser.identifierName(p.Variable.(*expr.Variable))
				v := b.HasVariable(vn)
				if v == nil || v.Type() == lang.Void {
					panic(vn + " is not defined.")
				}

				scalar := &lang.FunctionCall{
					Name:   "array.NewScalar",
					Args:   []lang.Expression{parser.expression(b, p.Dim)},
					Return: lang.String,
				}

				parser.gc.RequireNamespace("array")

				fc := &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.At", v),
					Args:   []lang.Expression{scalar},
					Return: ArrayItem(v.Type()),
				}
				scalar.SetParent(fc)
				fc.SetParent(b)

				// TODO: Type could know this.
				switch fc.Type() {
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

	case *expr.Isset:
		if len(e.Variables) != 1 {
			panic(`Isset can have only one argument, for now.`)
		}
		adf, ok := e.Variables[0].(*expr.ArrayDimFetch)
		if !ok {
			panic(`Only arrays are accepted for isset.`)
		}
		vn := parser.identifierName(adf.Variable.(*expr.Variable))
		v := b.HasVariable(vn)
		if v == nil || v.Type() == lang.Void {
			panic(vn + " is not defined.")
		}
		scalar := &lang.FunctionCall{
			Name:   "array.NewScalar",
			Args:   []lang.Expression{parser.expression(b, adf.Dim)},
			Return: lang.String,
		}

		parser.gc.RequireNamespace("array")

		fc := &lang.FunctionCall{
			Name:   fmt.Sprintf("%s.Isset", v),
			Args:   []lang.Expression{scalar},
			Return: lang.Bool,
		}

		fc.SetParent(b)
		return fc

	case *expr.UnaryPlus:
		expr := parser.expression(b, e.Expr)
		expr.SetParent(b)
		return expr

	case *expr.UnaryMinus:
		m := &lang.UnaryMinus{
			Expr: parser.expression(b, e.Expr),
		}
		m.Expr.SetParent(m)
		m.SetParent(b)
		return m

	case *scalar.Lnumber:
		n := &lang.Number{
			Value: e.Value,
		}
		n.SetParent(b)
		return n

	case *scalar.Dnumber:
		f := &lang.Float{
			Value: e.Value,
		}
		f.SetParent(b)
		return f

	case *scalar.String:
		s := e.Value
		if s[0] == '\'' && s[len(s)-1] == '\'' {
			s = strings.ReplaceAll(s, "\\", "\\\\")
			s = strings.ReplaceAll(s, "\"", "\\\"")
			s = strings.ReplaceAll(s, "'", "\"")
		} else if s[0] != '"' && s[len(s)-1] != '"' {
			s = fmt.Sprintf("\"%s\"", s)
		}
		str := &lang.Str{
			Value: s,
		}
		str.SetParent(b)
		return str

	case *expr.ShortArray:
		items := make([]lang.Expression, 0)
		for _, i := range e.Items {
			v := i.(*expr.ArrayItem).Val
			if v == nil {
				continue
			}
			items = append(items, parser.expression(b, v))
		}
		if len(items) == 0 {
			panic(`Cannot decide type, empty array.`)
		}
		typ := items[0].Type()
		for _, i := range items {
			if i.Type() != typ {
				panic(`Type is not the same for every element of the array.`)
			}
		}

		fc := &lang.FunctionCall{
			Name:   "array.New" + FirstUpper(typ),
			Args:   items,
			Return: ArrayType(typ),
		}

		parser.gc.RequireNamespace("array")

		fc.SetParent(b)
		return fc

	case *expr.ArrayDimFetch:
		v, ok := parser.expression(b, e.Variable).(*lang.VarRef)
		if !ok {
			panic(`Expected variable to be indexed.`)
		}

		scalar := &lang.FunctionCall{
			Name:   "array.NewScalar",
			Args:   []lang.Expression{parser.expression(b, e.Dim)},
			Return: lang.String,
		}

		parser.gc.RequireNamespace("array")

		fc := &lang.FunctionCall{
			Name:   fmt.Sprintf("%s.At", v),
			Args:   []lang.Expression{scalar},
			Return: ArrayItem(v.Type()),
		}
		scalar.SetParent(fc)
		fc.SetParent(b)
		return fc

	case *binary.Plus:
		return parser.binaryOp(b, "+", e.Left, e.Right)

	case *binary.Minus:
		return parser.binaryOp(b, "-", e.Left, e.Right)

	case *binary.Mul:
		return parser.binaryOp(b, "*", e.Left, e.Right)

	case *binary.Smaller:
		return parser.binaryOp(b, "<", e.Left, e.Right)

	case *binary.SmallerOrEqual:
		return parser.binaryOp(b, "<=", e.Left, e.Right)

	case *binary.GreaterOrEqual:
		return parser.binaryOp(b, ">=", e.Left, e.Right)

	case *binary.Greater:
		return parser.binaryOp(b, ">", e.Left, e.Right)

	case *binary.Identical:
		return parser.binaryOp(b, "==", e.Left, e.Right)

	case *expr.ConstFetch:
		c := &lang.Const{
			Value: parser.constructName(e.Constant.(*name.Name), true),
		}
		c.SetParent(b)
		return c

	// TODO: Add std functions to this parser, so it does not have to be
	// hacked like this.
	case *binary.Concat:
		f := &lang.FunctionCall{
			Name:   "std.Concat",
			Args:   make([]lang.Expression, 0),
			Return: lang.String,
		}
		f.AddArg(parser.expression(b, e.Left))
		f.AddArg(parser.expression(b, e.Right))
		parser.gc.RequireNamespace("std")
		return f

	case *expr.FunctionCall:
		al := e.ArgumentList

		n := parser.constructName(e.Function.(*name.Name), false)
		if ok := PHPFunctions[n]; ok {
			if n == "array_push" {
				if len(al.Arguments) < 2 {
					panic(`array_push requires atlast two arguments`)
				}
				v, ok := parser.expression(b, al.Arguments[0].(*node.Argument).Expr).(*lang.VarRef)
				if !ok || !IsArray(v.Type()) {
					panic(`First argument has to be a variable, an array.`)
				}
				typ := ArrayItem(v.Type())

				vars := []lang.Expression{}
				for _, arg := range al.Arguments[1:] {
					v := parser.expression(b, arg.(*node.Argument).Expr)
					if v.Type() != typ {
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
		n = parser.constructName(e.Function.(*name.Name), false)
		lf := parser.gc.Get(n)
		if lf == nil {
			panic(n + " is not defined")
		}

		f := &lang.FunctionCall{
			Name:   n,
			Args:   make([]lang.Expression, 0),
			Return: parser.gc.Get(n).Type(),
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
			arg := parser.expression(b, a.(*node.Argument).Expr)
			f.AddArg(arg)
		}
		return f
	}
	return nil
}

func (p *parser) binaryOp(b lang.Block, op string, left, right node.Node) lang.Expression {
	l := p.expression(b, left)
	r := p.expression(b, right)
	if convertToMatchingType(l, r) {
		p.gc.RequireNamespace("std")
	}
	res, err := lang.NewBinaryOp(op, l, r)
	if err != nil {
		panic(err)
	}
	res.SetParent(b)
	return res
}

// Returns a sign if namespace "std" has to be imported.
func convertToMatchingType(left, right lang.Expression) bool {
	lt := left.Type()
	rt := right.Type()
	if lt == rt {
		return false
	}

	// PHP tries to convert string to number,
	// skipping for now.
	t := false
	switch lt {
	case lang.Bool:
		switch rt {
		case lang.Int:
			f := &lang.FunctionCall{
				Name:   "std.BoolToInt",
				Args:   []lang.Expression{left},
				Return: lang.Int,
			}
			f.SetParent(left)
			left = f
			t = true

		case lang.Float64:
			f := &lang.FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   []lang.Expression{left},
				Return: lang.Float64,
			}
			f.SetParent(left)
			left = f
			t = true
		}

	case lang.Int:
		switch rt {
		case lang.Bool:
			f := &lang.FunctionCall{
				Name:   "std.BoolToInt",
				Args:   []lang.Expression{right},
				Return: lang.Int,
			}
			f.SetParent(right)
			right = f
			t = true

		case lang.Float64:
			f := &lang.FunctionCall{
				Name:   "float64",
				Args:   []lang.Expression{left},
				Return: lang.Float64,
			}
			f.SetParent(left)
			left = f
			t = true
		}

	case lang.Float64:
		switch rt {
		case lang.Bool:
			f := &lang.FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   []lang.Expression{right},
				Return: lang.Float64,
			}
			f.SetParent(right)
			right = f
			t = true

		case lang.Int:
			f := &lang.FunctionCall{
				Name:   "float",
				Args:   []lang.Expression{right},
				Return: lang.Float64,
			}
			f.SetParent(right)
			right = f
			t = true
		}
	}
	return t
}

// conditionExpr lf, f.Cond[0parses next expression and if it does not
// return bool, adds function std.Truthy(interface{})
func (p *parser) conditionExpr(b lang.Block, n node.Node) lang.Expression {
	e := p.expression(b, n)
	if e == nil {
		panic(`conditionExpr: missing expression`)
	}
	if e.Type() != lang.Bool {
		e = &lang.FunctionCall{
			Name:   "std.Truthy",
			Args:   []lang.Expression{e},
			Return: lang.Bool,
		}
		e.SetParent(b)
		p.gc.RequireNamespace("std")
	}

	return e
}

func checkArguments(vars []*lang.Variable, call []node.Node) error {
	if len(vars) != len(call) {
		return errors.New("wrong argument count")
	}
	// TODO: Check if arguments are passed by reference, make sure
	// that is done only with variables.
	return nil
}

func (parser *parser) buildAssignment(parent lang.Block, name string, right lang.Expression) *lang.Assign {
	t := right.Type()
	if t == lang.Void {
		panic("Cannot assign \"void\" " + "to \"" + name + "\".")
	}

	if parser.useGlobalContext {
		name = parser.translator.Translate(name, Public)
	} else {
		name = parser.translator.Translate(name, Private)
	}

	v := parent.HasVariable(name)
	fd := false
	if v == nil {
		v = parser.newVariable(name, t, false)
		if parser.useGlobalContext {
			parser.gc.DefineVariable(v)
		} else {
			parent.DefineVariable(v)
		}
		fd = true
	} else if v.CurrentType != t {
		if parser.useGlobalContext {
			parser.gc.DefineVariable(v)
			v.CurrentType = t
		} else {
			if v.FirstDefinition.Parent() == parent {
				v.CurrentType = t
			} else {
				v = parser.newVariable(name, t, false)
				fd = true
			}
		}
		parent.DefineVariable(v)
	}

	as, err := lang.NewAssign(v, right)
	if err != nil {
		panic(err)
	}

	if parser.useGlobalContext && fd {
		v.FirstDefinition = parser.gc
		fd = false
	}
	if fd {
		v.FirstDefinition = as
	}

	as.FirstDefinition = fd
	as.SetParent(parent)

	return as
}

func (p *parser) newVariable(name, typ string, isConst bool) *lang.Variable {
	// TODO: This fixed "Public" does not look right.
	// It should probably require already unique name.
	name = p.translator.Translate(name, Public)
	return lang.NewVariable(name, typ, isConst)
}

/**
 * Function makes things much easier, I expect
 * identifier name to be just simple right now
 * defined string, no variable etc.
 */
func (p *parser) identifierName(v *expr.Variable) string {
	switch v.VarName.(type) {
	case *node.Identifier:
		n := v.VarName.(*node.Identifier).Value
		if p.useGlobalContext {
			return p.translator.Translate(n, Public)
		} else {
			return p.translator.Translate(n, Private)
		}

	default:
		panic(`Variable name is not defined as a simple string.`)
	}
}

func (p *parser) constructName(nm *name.Name, translate bool) string {
	s := ""
	for _, n := range nm.Parts {
		s += n.(*name.NamePart).Value
	}
	if !translate {
		return s
	}
	if p.useGlobalContext {
		return p.functionTranslator.Translate(s, Public)
	}
	return p.functionTranslator.Translate(s, Private)
}
