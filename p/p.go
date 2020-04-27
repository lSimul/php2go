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
		switch s.(type) {
		case *stmt.Function:
			functions = append(functions, *s.(*stmt.Function))
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
			parser.createFunction(list, s.(*stmt.StmtList).Stmts)

		case *stmt.Expression:
			ex := parser.makeExpression(b, s.(*stmt.Expression))
			b.AddStatement(ex)

		case *stmt.For:
			f := s.(*stmt.For)
			lf := lang.ConstructFor(b)

			if f.Init != nil {
				n := f.Init[0]
				ex := parser.simpleExpression(lf, n)
				lf.Init = ex
			}

			if f.Cond != nil {
				n := f.Cond[0]
				e := parser.expression(lf, n)
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
				ex := parser.simpleExpression(lf, n)
				lf.Loop = ex
			}

			parser.createFunction(lf.Block, nodeList(f.Stmt))
			b.AddStatement(lf)

		case *stmt.While:
			w := s.(*stmt.While)
			lf := lang.ConstructFor(b)

			e := parser.expression(lf, w.Cond)
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

			parser.createFunction(lf.Block, nodeList(w.Stmt))
			b.AddStatement(lf)

		case *stmt.Do:
			w := s.(*stmt.Do)
			lf := lang.ConstructFor(b)

			parser.createFunction(lf.Block, nodeList(w.Stmt))

			i := &lang.If{
				Vars: make([]*lang.Variable, 0),
			}
			i.True = lang.NewCode(i)
			i.SetParent(lf)
			// TODO: Negation should work only with
			// boolean values.
			c := parser.expression(i, w.Cond)
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
			lf.Block = lang.NewCode(lf)

			iterated := parser.expression(lf, f.Expr)
			if !IsArray(iterated.GetType()) {
				panic(`Only arrays can be iterated.`)
			}

			var fnName string
			switch iterated.(type) {
			case *lang.VarRef:
				fnName = iterated.(*lang.VarRef).V.Name
			case *lang.FunctionCall:
				fnName = iterated.(*lang.FunctionCall).String()

			default:
				panic(`Uncatched type of iterated.`)
			}

			var it *lang.FunctionCall
			// Easy chain, without array.Pair
			if f.Key == nil {
				it = &lang.FunctionCall{
					Name: fnName + ".Iter",
					// TODO: Set up return type.
				}
				name := parser.identifierName(f.Variable.(*expr.Variable))
				lf.Value = *parser.newVariable(name, ArrayItem(iterated.GetType()), false)
			} else {
				name := parser.identifierName(f.Key.(*expr.Variable))
				// TODO: Define type scalar which is being
				// formated as string.
				k := parser.newVariable(name, lang.String, false)
				n := parser.identifierName(f.Variable.(*expr.Variable))
				v := parser.newVariable(n, ArrayItem(iterated.GetType()), false)

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
					s, err := lang.NewAssign(k, lang.NewVarRef(pairK, pairK.GetType()))
					if err != nil {
						panic(err)
					}
					s.FirstDefinition = true
					lf.Block.AddStatement(s)
					lf.Block.DefineVariable(k)
				}

				pairV := parser.newVariable(lf.Value.Name+".V", ArrayItem(iterated.GetType()), true)
				s, err := lang.NewAssign(v, lang.NewVarRef(pairV, pairV.GetType()))
				if err != nil {
					panic(err)
				}
				s.FirstDefinition = true
				lf.Block.AddStatement(s)
				lf.Block.DefineVariable(v)
			}
			it.SetParent(lf)
			lf.Iterated = it

			parser.createFunction(lf.Block, nodeList(f.Stmt))

			lf.SetParent(b)
			b.AddStatement(lf)

		case *stmt.If:
			i := parser.constructIf(b, s.(*stmt.If))
			b.AddStatement(i)

		case *stmt.Switch:
			s := s.(*stmt.Switch)
			sw := &lang.Switch{
				Cases: make([]lang.Node, 0),
			}
			sw.SetParent(b)
			sw.Condition = parser.expression(sw, s.Cond)
			parser.constructSwitch(sw, s.CaseList)
			b.AddStatement(sw)

		case *stmt.Return:
			r := &lang.Return{
				Expression: parser.expression(b, s.(*stmt.Return).Expr),
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
	expr := p.expression(nif, i.Cond)
	if expr == nil {
		panic(`constructIf: missing expression`)
	}
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
	switch e.(type) {
	case *stmt.If:
		lif.False = p.constructIf(lif, e.(*stmt.If))

	default:
		c := lang.NewCode(lif)
		p.createFunction(c, nodeList(e))
		lif.False = c
	}
	return nif
}

func (p *parser) constructElif(b lang.Block, i *stmt.ElseIf) *lang.If {
	nif := &lang.If{}
	nif.SetParent(b)
	e := p.expression(nif, i.Cond)
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
	p.createFunction(nif.True, nodeList(i.Stmt))
	return nif
}

func (parser *parser) constructSwitch(s *lang.Switch, cl *stmt.CaseList) {
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
			c := c.(*stmt.Default)
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

	switch n.(type) {
	case *assign.Assign:
		a := n.(*assign.Assign)

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

	switch n.(type) {
	// Every expression should have return value.
	// Otherwise I cannot say what the assigned value will have.
	case *assign.Assign:
		a := n.(*assign.Assign)

		r := parser.complexExpression(b, a.Expression)
		if r == nil {
			panic(`Missing right side for assignment.`)
		}

		la, ok := r.(*lang.Assign)
		if ok {
			b.AddStatement(la)
			r = lang.NewVarRef(la.Left(), la.GetType())
		}

		switch a.Variable.(type) {
		case (*expr.Variable):
			v := a.Variable.(*expr.Variable)
			n := parser.identifierName(v)
			return parser.buildAssignment(b, n, r)

		case (*expr.ArrayDimFetch):
			adf := a.Variable.(*expr.ArrayDimFetch)
			vn := parser.identifierName(adf.Variable.(*expr.Variable))
			v := b.HasVariable(vn)
			if v == nil || v.GetType() == lang.Void {
				panic(vn + " is not defined.")
			}

			if l, r := ArrayItem(v.GetType()), r.GetType(); l != r {
				panic(fmt.Sprintf("Array editing: '%s' expected, '%s' given.", l, r))
			}

			var fc *lang.FunctionCall
			if adf.Dim == nil {
				fc = &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.Add", v),
					Args:   []lang.Expression{r},
					Return: v.GetType(),
				}
				fc.SetParent(b)
			} else {
				scalar := &lang.FunctionCall{
					Name:   "array.NewScalar",
					Args:   []lang.Expression{parser.expression(b, adf.Dim)},
					Return: lang.String,
				}
				fc = &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.Edit", v),
					Args:   []lang.Expression{scalar, r},
					Return: v.GetType(),
				}
				scalar.SetParent(fc)
				fc.SetParent(b)
			}
			return fc

		default:
			panic(fmt.Sprintf("Unexpected left side: %v", a))
		}

	}
	panic(`ComplexExpression: something else uncatched.`)
}

func (parser *parser) statement(b lang.Block, n node.Node) lang.Node {
	// Inpisration from parser.expression, *expr.ArrayDimFetch.
	switch n.(type) {
	case *stmt.Unset:
		first := true
		var ret lang.Node
		unsets := n.(*stmt.Unset)
		for _, v := range unsets.Vars {
			adf, ok := v.(*expr.ArrayDimFetch)
			if !ok {
				panic(`Only arrays are accepted for unset.`)
			}
			vn := parser.identifierName(adf.Variable.(*expr.Variable))
			v := b.HasVariable(vn)
			if v == nil || v.GetType() == lang.Void {
				panic(vn + " is not defined.")
			}
			scalar := &lang.FunctionCall{
				Name:   "array.NewScalar",
				Args:   []lang.Expression{parser.expression(b, adf.Dim)},
				Return: lang.String,
			}
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

	switch n.(type) {
	case *expr.PreInc:
		v, ok = parser.expression(b, n.(*expr.PreInc).Variable).(*lang.VarRef)
		return inc()
	case *expr.PostInc:
		v, ok = parser.expression(b, n.(*expr.PostInc).Variable).(*lang.VarRef)
		return inc()

	case *expr.PreDec:
		v, ok = parser.expression(b, n.(*expr.PreDec).Variable).(*lang.VarRef)
		return dec()
	case *expr.PostDec:
		v, ok = parser.expression(b, n.(*expr.PostDec).Variable).(*lang.VarRef)
		return dec()
	}
	return nil
}

func (parser *parser) expression(b lang.Block, n node.Node) lang.Expression {
	switch n.(type) {
	case *expr.Variable:
		name := parser.identifierName(n.(*expr.Variable))
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
				vn := parser.identifierName(p.(*expr.Variable))
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
				vn := parser.identifierName(adf.Variable.(*expr.Variable))
				v := b.HasVariable(vn)
				if v == nil || v.GetType() == lang.Void {
					panic(vn + " is not defined.")
				}

				scalar := &lang.FunctionCall{
					Name:   "array.NewScalar",
					Args:   []lang.Expression{parser.expression(b, adf.Dim)},
					Return: lang.String,
				}
				fc := &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.At", v),
					Args:   []lang.Expression{scalar},
					Return: ArrayItem(v.GetType()),
				}
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

	case *expr.Isset:
		issets := n.(*expr.Isset)
		if len(issets.Variables) != 1 {
			panic(`Isset can have only one argument, for now.`)
		}
		adf, ok := issets.Variables[0].(*expr.ArrayDimFetch)
		if !ok {
			panic(`Only arrays are accepted for isset.`)
		}
		vn := parser.identifierName(adf.Variable.(*expr.Variable))
		v := b.HasVariable(vn)
		if v == nil || v.GetType() == lang.Void {
			panic(vn + " is not defined.")
		}
		scalar := &lang.FunctionCall{
			Name:   "array.NewScalar",
			Args:   []lang.Expression{parser.expression(b, adf.Dim)},
			Return: lang.String,
		}
		fc := &lang.FunctionCall{
			Name: fmt.Sprintf("%s.Isset", v),
			Args: []lang.Expression{scalar},
		}

		fc.SetParent(b)
		return fc

	case *expr.UnaryPlus:
		e := parser.expression(b, n.(*expr.UnaryPlus).Expr)
		e.SetParent(b)
		return e

	case *expr.UnaryMinus:
		m := &lang.UnaryMinus{
			Expr: parser.expression(b, n.(*expr.UnaryMinus).Expr),
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
		} else if s[0] != '"' && s[len(s)-1] != '"' {
			s = fmt.Sprintf("\"%s\"", s)
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
			items = append(items, parser.expression(b, v))
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
		v, ok := parser.expression(b, adf.Variable).(*lang.VarRef)
		if !ok {
			panic(`Expected variable to be indexed.`)
		}

		scalar := &lang.FunctionCall{
			Name:   "array.NewScalar",
			Args:   []lang.Expression{parser.expression(b, adf.Dim)},
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
		return parser.binaryOp(b, "+", p.Left, p.Right)

	case *binary.Minus:
		p := n.(*binary.Minus)
		return parser.binaryOp(b, "-", p.Left, p.Right)

	case *binary.Mul:
		p := n.(*binary.Mul)
		return parser.binaryOp(b, "*", p.Left, p.Right)

	case *binary.Smaller:
		p := n.(*binary.Smaller)
		return parser.binaryOp(b, "<", p.Left, p.Right)

	case *binary.SmallerOrEqual:
		p := n.(*binary.SmallerOrEqual)
		return parser.binaryOp(b, "<=", p.Left, p.Right)

	case *binary.GreaterOrEqual:
		p := n.(*binary.GreaterOrEqual)
		return parser.binaryOp(b, ">=", p.Left, p.Right)

	case *binary.Greater:
		p := n.(*binary.Greater)
		return parser.binaryOp(b, ">", p.Left, p.Right)

	case *binary.Identical:
		p := n.(*binary.Identical)
		return parser.binaryOp(b, "==", p.Left, p.Right)

	case *expr.ConstFetch:
		cf := n.(*expr.ConstFetch)
		c := &lang.Const{
			Value: parser.constructName(cf.Constant.(*name.Name), true),
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
		f.AddArg(parser.expression(b, c.Left))
		f.AddArg(parser.expression(b, c.Right))
		return f

	case *expr.FunctionCall:
		fc := n.(*expr.FunctionCall)
		al := fc.ArgumentList

		n := parser.constructName(fc.Function.(*name.Name), false)
		if ok := PHPFunctions[n]; ok {
			if n == "array_push" {
				if len(al.Arguments) < 2 {
					panic(`array_push requires atlast two arguments`)
				}
				v, ok := parser.expression(b, al.Arguments[0].(*node.Argument).Expr).(*lang.VarRef)
				if !ok {
					panic(`First argument has to be a variable.`)
				}
				typ := ArrayItem(v.GetType())

				vars := []lang.Expression{}
				for _, arg := range al.Arguments[1:] {
					v := parser.expression(b, arg.(*node.Argument).Expr)
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
		n = parser.constructName(fc.Function.(*name.Name), false)
		lf := parser.gc.Get(n)
		if lf == nil {
			panic(n + " is not defined")
		}

		f := &lang.FunctionCall{
			Name:   n,
			Args:   make([]lang.Expression, 0),
			Return: parser.gc.Get(n).GetType(),
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
	res, err := lang.NewBinaryOp(op, p.expression(b, left), p.expression(b, right))
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

func (parser *parser) buildAssignment(parent lang.Block, name string, right lang.Expression) *lang.Assign {
	t := right.GetType()
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
			// TODO: This is something I solve in the define variable.
			v.Type = lang.Anything
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
