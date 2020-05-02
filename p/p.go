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

	asServer bool

	gc    *lang.GlobalContext
	funcs *Func
}

func NewParser(v, f NameTranslation) *parser {
	return &parser{
		translator:         v,
		functionTranslator: f,
		asServer:           false,
	}
}

func (p *parser) Run(r *node.Root, asServer bool) *lang.GlobalContext {
	p.gc = lang.NewGlobalContext()
	p.funcs = NewFunc(p.gc)
	if asServer {
		p.globalContextAsServer()
	}
	ms, fs := sanitizeRootStmts(r)

	for _, s := range fs {
		f, defaultParams := p.funcDef(&s)
		p.funcs.Add(f.Name, f, 0)

		for i := len(defaultParams) - 1; i >= 0; i-- {
			n := p.functionTranslator.Translate(fmt.Sprintf("%s%d", f.Name, i))
			vf := lang.NewFunc(n)
			var args []lang.Expression
			for j := 0; j < len(f.Args)-(len(defaultParams)-i); j++ {
				v := lang.NewVariable(f.Args[j].Name, f.Args[j].Type(), false)
				vf.Args = append(vf.Args, v)
				args = append(args, lang.NewVarRef(v, v.CurrentType))
			}
			args = append(args, defaultParams[i])

			c, err := p.funcs.Namespace("").Call(f.Name, args)
			if err != nil {
				panic(err)
			}
			if f.Return.Equal(lang.Void) {
				vf.Body.AddStatement(c)
			} else {
				vf.Body.AddStatement(&lang.Return{Expression: c})
			}
			p.funcs.Add(f.Name, vf, len(defaultParams)-i)
		}
	}
	for _, s := range fs {
		f, _ := p.funcDef(&s)
		p.createFunction(&f.Body, s.Stmts)
		p.funcs.Add(f.Name, f, 0)
	}

	main := p.mainDef()
	p.funcs.Add(main.Name, main, 0)
	p.createFunction(&main.Body, ms)
	return p.gc
}

func (p *parser) globalContextAsServer() {
	p.asServer = true
	p.gc.AddImport("flag")
	p.gc.AddImport("log")
	p.gc.AddImport("net/http")
	p.gc.AddImport("os")

	p.gc.AddImport("io")
	p.gc.DefineVariable(lang.NewVariable("W", lang.NewTyp(lang.Writer, false), false))

	p.funcs.Namespace("array")
	p.gc.DefineVariable(lang.NewVariable("_GET", lang.NewTyp(ArrayType(lang.String), false), false))
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

func (parser *parser) funcDef(fc *stmt.Function) (*lang.Function, []lang.Expression) {
	defaultParams := make([]lang.Expression, 0)
	if fc == nil {
		return nil, defaultParams
	}

	n := parser.functionTranslator.Translate(fc.FunctionName.(*node.Identifier).Value)
	f := lang.NewFunc(n)
	f.SetParent(parser.gc)

	hasDefaultParams := false
	for _, pr := range fc.Params {
		p := pr.(*node.Parameter)
		// Default value has to be by value.
		typ := lang.NewTyp(parser.constructName(p.VariableType.(*name.Name), true), p.ByRef)
		v := lang.NewVariable(
			parser.identifierName(p.Variable.(*expr.Variable)),
			typ, false)

		if p.DefaultValue != nil {
			dv := parser.expression(nil, p.DefaultValue)
			defaultParams = append(defaultParams, dv)
			hasDefaultParams = true
		} else if hasDefaultParams {
			panic(`Default parameters cannot be defined with gaps.`)
		}

		f.Args = append(f.Args, v)
	}

	if fc.ReturnType != nil {
		n := parser.constructName(fc.ReturnType.(*name.Name), true)
		if n == "void" {
			n = lang.Void
		}
		// In PHP every return type is by value.
		f.Return = lang.NewTyp(n, false)
	}

	return f, defaultParams
}

func (p *parser) mainDef() *lang.Function {
	var f *lang.Function
	if p.asServer {
		f = lang.NewFunc("mainFunc")
	} else {
		f = lang.NewFunc("main")
	}
	f.SetParent(p.gc)
	return f
}

func (parser *parser) createFunction(b lang.Block, stmts []node.Node) {
	for _, s := range stmts {
		switch s := s.(type) {
		case *stmt.Nop:
			// Alias for <?php ?> and "empty semicolon", nothing to do.

		case *stmt.InlineHtml:
			str := &lang.Str{Value: fmt.Sprintf("`%s`", s.Value)}
			var err error
			var f *lang.FunctionCall
			if parser.asServer {
				f, err = parser.servePrint([]lang.Expression{str})
			} else {
				f, err = parser.funcs.Namespace("fmt").Call("Print", []lang.Expression{str})
			}
			if err != nil {
				panic(err)
			}
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
			lf := lang.NewFor(b)

			if s.Init != nil {
				n := s.Init[0]
				ex := parser.simpleExpression(lf, n)
				lf.Init = ex
			}

			if s.Cond != nil {
				loop, cond := parser.flowControlExpr(lf, s.Cond[0])

				if loop != nil {
					lf.Block.AddStatement(loop)
				}
				err := lf.SetCond(cond)
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
			lf := lang.NewFor(b)
			loop, cond := parser.flowControlExpr(lf, s.Cond)

			if loop != nil {
				lf.Block.AddStatement(loop)
			}
			err := lf.SetCond(cond)
			if err != nil {
				panic(err)
			}

			parser.createFunction(lf.Block, nodeList(s.Stmt))
			b.AddStatement(lf)

		case *stmt.Do:
			lf := lang.NewFor(b)

			parser.createFunction(lf.Block, nodeList(s.Stmt))

			i := lang.NewIf(lf)
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
			if !IsArray(iterated.Type().String()) {
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
				typ := lang.NewTyp(ArrayItem(iterated.Type().String()), false)
				lf.Value = *lang.NewVariable(name, typ, false)
			} else {
				name := parser.identifierName(s.Key.(*expr.Variable))
				// TODO: Define type scalar which is being
				// formated as string.
				k := lang.NewVariable(name, lang.NewTyp(lang.String, false), false)
				n := parser.identifierName(s.Variable.(*expr.Variable))
				typ := lang.NewTyp(ArrayItem(iterated.Type().String()), false)
				v := lang.NewVariable(n, typ, false)

				it = &lang.FunctionCall{
					Name: fnName + ".KeyIter",
					// TODO: Set up return type.
				}

				// TODO: This is not lang.Void
				lf.Value = *lang.NewVariable("pair", lang.NewTyp(lang.Void, false), false)

				// TODO: I do not have this part of code under control.
				// Accessing struct elements is out of my reach right now.
				if k != nil {
					pairK := lang.NewVariable(lf.Value.Name+".K", lang.NewTyp(lang.String, false), true)
					s, err := lang.NewAssign(k, lang.NewVarRef(pairK, pairK.Type()))
					if err != nil {
						panic(err)
					}
					s.FirstDefinition = true
					lf.Block.AddStatement(s)
					lf.Block.DefineVariable(k)
				}

				typ = lang.NewTyp(ArrayItem(iterated.Type().String()), false)
				pairV := lang.NewVariable(lf.Value.Name+".V", typ, true)
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
			r := &lang.Return{}
			if s.Expr != nil {
				r.Expression = parser.expression(b, s.Expr)
			}
			b.AddStatement(r)

		case *stmt.Echo:
			var args []lang.Expression
			for _, e := range s.Exprs {
				// TODO: Do not ignore information in Argument,
				// it has interesting information like if it is
				// send by reference and others.
				args = append(args, parser.expression(b, e))
			}

			var err error
			var f *lang.FunctionCall
			if parser.asServer {
				f, err = parser.servePrint(args)
			} else {
				f, err = parser.funcs.Namespace("fmt").Call("Print", args)
			}
			if err != nil {
				panic(err)
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
	}
	if n == nil {
		return []node.Node{}
	} else {
		return []node.Node{n}
	}
}

func (p *parser) constructIf(b lang.Block, i *stmt.If) *lang.If {
	nif := lang.NewIf(b)
	init, cond := p.flowControlExpr(nif, i.Cond)

	nif.Init = init
	err := nif.SetCond(cond)
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
	nif := lang.NewIf(b)
	init, cond := p.flowControlExpr(nif, i.Cond)

	nif.Init = init
	err := nif.SetCond(cond)
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
			lc := lang.NewCase(s)
			s.Cases = append(s.Cases, lc)
			lc.Condition = parser.expression(lc, c.Cond)
			parser.createFunction(lc.Block, c.Stmts)
			b := lc.Block
			if len(b.Statements) <= 0 {
				break
			}
			if _, ok := b.Statements[len(b.Statements)-1].(*lang.Break); ok {
				b.Statements = b.Statements[:len(b.Statements)-1]
			} else {
				f := &lang.Fallthrough{}
				f.SetParent(lc)
				b.Statements = append(b.Statements, f)
			}

		case *stmt.Default:
			d := lang.NewDefault(s)
			s.Cases = append(s.Cases, d)
			parser.createFunction(d.Block, c.Stmts)

			b := d.Block
			if len(b.Statements) <= 0 {
				break
			}

			_, ok := b.Statements[len(b.Statements)-1].(*lang.Break)
			if ok {
				b.Statements = b.Statements[:len(b.Statements)-1]
			} else {
				f := &lang.Fallthrough{}
				f.SetParent(d)
				b.Statements = append(b.Statements, f)
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

	a := p.directAssignment(b, n)
	if a != nil {
		return a
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

	if a := parser.directAssignment(b, n); a != nil {
		return a
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
		vr := b.HasVariable(vn, true)
		if vr == nil || vr.Type().Equal(lang.Void) {
			panic(vn + " is not defined.")
		}

		if l, r := ArrayItem(vr.Type().String()), r.Type().String(); l != r {
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
			args := []lang.Expression{parser.expression(b, v.Dim)}
			scalar, err := parser.funcs.Namespace("array").Call("NewScalar", args)
			if err != nil {
				panic(err)
			}
			fc = &lang.FunctionCall{
				Name:   fmt.Sprintf("%s.Edit", vr),
				Args:   []lang.Expression{scalar, r},
				Return: vr.Type(),
			}

			scalar.SetParent(fc)
			fc.SetParent(b)
		}
		return fc

	default:
		panic(fmt.Sprintf("Unexpected left side: %v", a))
	}
}

func (parser *parser) directAssignment(b lang.Block, n node.Node) lang.Expression {
	assignmentFunc := func(op string, expr node.Node, nv *expr.Variable) *lang.Assign {
		e := parser.expression(b, expr)
		n := parser.identifierName(nv)
		v := b.HasVariable(n, true)
		if v == nil {
			panic(n + " is not defined.")
		}
		e = parser.bOp(b, op, lang.NewVarRef(v, v.CurrentType), e)
		if e == nil {
			panic("Issue with a binary operand.")
		}
		return parser.buildAssignment(b, n, e)
	}

	switch a := n.(type) {
	case (*assign.Concat):
		e := parser.expression(b, a.Expression)
		n := parser.identifierName(a.Variable.(*expr.Variable))
		v := b.HasVariable(n, true)
		if v == nil {
			panic(n + " is not defined.")
		}
		fc, err := parser.funcs.Namespace("std").Call("Concat", []lang.Expression{
			lang.NewVarRef(v, v.CurrentType), e,
		})
		if err != nil {
			panic(err)
		}
		return parser.buildAssignment(b, n, fc)

	case (*assign.Plus):
		return assignmentFunc("+", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.Minus):
		return assignmentFunc("-", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.Div):
		return assignmentFunc("/", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.Mul):
		return assignmentFunc("*", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.Mod):
		return assignmentFunc("%", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.BitwiseAnd):
		return assignmentFunc("&", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.BitwiseOr):
		return assignmentFunc("|", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.BitwiseXor):
		return assignmentFunc("^", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.ShiftLeft):
		return assignmentFunc("<<", a.Expression, a.Variable.(*expr.Variable))

	case (*assign.ShiftRight):
		return assignmentFunc(">>", a.Expression, a.Variable.(*expr.Variable))
	}

	return nil
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
			v := b.HasVariable(vn, true)
			if v == nil || v.Type().Equal(lang.Void) {
				panic(vn + " is not defined.")
			}
			args := []lang.Expression{parser.expression(b, adf.Dim)}
			scalar, err := parser.funcs.Namespace("array").Call("NewScalar", args)
			if err != nil {
				panic(err)
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
		if b.HasVariable(v.V.Name, true) == nil {
			panic(fmt.Sprintf("'%s' is not defined.", v.V.Name))
		}
		return lang.NewInc(
			b, v,
			func(e lang.Expression) (lang.Expression, error) {
				return parser.funcs.Namespace("std").Call("StrInc", []lang.Expression{e})
			},
		)
	}
	dec := func() lang.Node {
		if !ok {
			panic(`"--" requires variable.`)
		}
		if b.HasVariable(v.V.Name, true) == nil {
			panic(fmt.Sprintf("'%s' is not defined.", v.V.Name))
		}
		return lang.NewDec(
			b, v,
			func(e lang.Expression) (lang.Expression, error) {
				return parser.funcs.Namespace("std").Call("StrDec", []lang.Expression{e})
			},
		)
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
		v := b.HasVariable(name, true)
		if v == nil {
			panic("Using undefined variable \"" + name + "\".")
		}
		return lang.NewVarRef(v, v.Type())

	case *scalar.Encapsed:
		args := make([]lang.Expression, 1)
		s := &lang.Str{
			Value: "\"",
		}
		for _, p := range e.Parts {
			switch p := p.(type) {
			case *scalar.EncapsedStringPart:
				s.Value += p.Value

			case *expr.Variable:
				vn := parser.identifierName(p)
				v := b.HasVariable(vn, true)
				if v == nil || v.Type().Equal(lang.Void) {
					panic(vn + " is not defined.")
				}
				s.Value += varFormat(v.Type().String())
				args = append(args, lang.NewVarRef(v, v.Type()))

			case *expr.ArrayDimFetch:
				vn := parser.identifierName(p.Variable.(*expr.Variable))
				v := b.HasVariable(vn, true)
				if v == nil || v.Type().Equal(lang.Void) {
					panic(vn + " is not defined.")
				}

				scalar, err := parser.funcs.Namespace("array").Call(
					"NewScalar", []lang.Expression{parser.expression(b, p.Dim)})
				if err != nil {
					panic(err)
				}

				fc := &lang.FunctionCall{
					Name:   fmt.Sprintf("%s.At", v),
					Args:   []lang.Expression{scalar},
					Return: lang.NewTyp(ArrayItem(v.Type().String()), false),
				}
				scalar.SetParent(fc)
				fc.SetParent(b)

				s.Value += varFormat(fc.Type().String())

				args = append(args, fc)
			}
		}
		s.Value += "\""
		args[0] = s

		f, err := parser.funcs.Namespace("fmt").Call("Sprintf", args)
		if err != nil {
			panic(err)
		}
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
		v := b.HasVariable(vn, true)
		if v == nil || v.Type().Equal(lang.Void) {
			panic(vn + " is not defined.")
		}

		args := []lang.Expression{parser.expression(b, adf.Dim)}
		scalar, err := parser.funcs.Namespace("array").Call("NewScalar", args)
		if err != nil {
			panic(err)
		}

		fc := &lang.FunctionCall{
			Name:   fmt.Sprintf("%s.Isset", v),
			Args:   []lang.Expression{scalar},
			Return: lang.NewTyp(lang.Bool, false),
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
		s = strings.ReplaceAll(s, "\\$", "$")
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
		typ := items[0].Type().String()
		for _, i := range items {
			if !i.Type().Equal(typ) {
				panic(`Type is not the same for every element of the array.`)
			}
		}

		f, err := parser.funcs.Namespace("array").Call("New"+FirstUpper(typ), items)
		if err != nil {
			panic(err)
		}

		f.SetParent(b)
		return f

	case *expr.ArrayDimFetch:
		v, ok := parser.expression(b, e.Variable).(*lang.VarRef)
		if !ok {
			panic(`Expected variable to be indexed.`)
		}

		args := []lang.Expression{parser.expression(b, e.Dim)}
		scalar, err := parser.funcs.Namespace("array").Call("NewScalar", args)
		if err != nil {
			panic(err)
		}

		fc := &lang.FunctionCall{
			Name:   fmt.Sprintf("%s.At", v),
			Args:   []lang.Expression{scalar},
			Return: lang.NewTyp(ArrayItem(v.Type().String()), false),
		}
		scalar.SetParent(fc)
		fc.SetParent(b)
		return fc

		/*
			TODO:
			[.] coalesce
			[.] equal
			[.] not_equal
			[.] pow // **
			[.] spaceship
		*/

	case *binary.Plus:
		return parser.binaryOp(b, "+", e.Left, e.Right)

	case *binary.Minus:
		return parser.binaryOp(b, "-", e.Left, e.Right)

	case *binary.Mul:
		return parser.binaryOp(b, "*", e.Left, e.Right)

	case *binary.Div:
		return parser.binaryOp(b, "/", e.Left, e.Right)

	case *binary.Mod:
		return parser.binaryOp(b, "%", e.Left, e.Right)

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

	case *binary.NotIdentical:
		return parser.binaryOp(b, "!=", e.Left, e.Right)

	case *binary.LogicalAnd:
		// and
		return parser.binaryOp(b, "&&", e.Left, e.Right)

	case *binary.BooleanAnd:
		return parser.binaryOp(b, "&&", e.Left, e.Right)

	case *binary.LogicalOr:
		// or
		return parser.binaryOp(b, "||", e.Left, e.Right)

	case *binary.BooleanOr:
		return parser.binaryOp(b, "||", e.Left, e.Right)

	case *binary.LogicalXor:
		// xor
		return parser.binaryOp(b, "^", e.Left, e.Right)

	case *binary.BitwiseXor:
		return parser.binaryOp(b, "^", e.Left, e.Right)

	case *binary.BitwiseAnd:
		return parser.binaryOp(b, "&", e.Left, e.Right)

	case *binary.BitwiseOr:
		return parser.binaryOp(b, "|", e.Left, e.Right)

	case *binary.ShiftLeft:
		return parser.binaryOp(b, "<<", e.Left, e.Right)

	case *binary.ShiftRight:
		return parser.binaryOp(b, ">>", e.Left, e.Right)

	case *expr.ConstFetch:
		c := &lang.Const{
			Value: parser.constructName(e.Constant.(*name.Name), true),
		}
		c.SetParent(b)
		return c

	// TODO: Add std functions to this parser, so it does not have to be
	// hacked like this.
	case *binary.Concat:
		args := []lang.Expression{
			parser.expression(b, e.Left),
			parser.expression(b, e.Right),
		}
		f, err := parser.funcs.Namespace("std").Call("Concat", args)
		if err != nil {
			panic(err)
		}
		return f

	case *expr.FunctionCall:
		n := parser.constructName(e.Function.(*name.Name), false)
		arguments := e.ArgumentList.Arguments
		args := make([]lang.Expression, 0, len(arguments))
		for _, a := range arguments {
			// TODO: Do not ignore information in Argument,
			// it has interesting information like if it is
			// send by reference and others.
			args = append(args, parser.expression(b, a.(*node.Argument).Expr))
		}

		if fc, ok := PHPFunctions[n]; ok {
			f, err := fc(b, args)
			if err != nil {
				panic(err)
			}
			return f
		}

		n = parser.constructName(e.Function.(*name.Name), true)

		f, err := parser.funcs.Namespace("").Call(n, args)
		if err != nil {
			panic(err)
		}
		return f
	}
	return nil
}

func (p *parser) binaryOp(b lang.Block, op string, left, right node.Node) lang.Expression {
	l := p.expression(b, left)
	r := p.expression(b, right)
	return p.bOp(b, op, l, r)
}

func (p *parser) bOp(b lang.Block, op string, l, r lang.Expression) lang.Expression {
	if convertToMatchingType(l, r) {
		p.funcs.Namespace("std")
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
	switch lt.String() {
	case lang.Bool:
		switch rt.String() {
		case lang.Int:
			f := &lang.FunctionCall{
				Name:   "std.BoolToInt",
				Args:   []lang.Expression{left},
				Return: lang.NewTyp(lang.Int, false),
			}
			f.SetParent(left)
			left = f
			t = true

		case lang.Float64:
			f := &lang.FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   []lang.Expression{left},
				Return: lang.NewTyp(lang.Float64, false),
			}
			f.SetParent(left)
			left = f
			t = true
		}

	case lang.Int:
		switch rt.String() {
		case lang.Bool:
			f := &lang.FunctionCall{
				Name:   "std.BoolToInt",
				Args:   []lang.Expression{right},
				Return: lang.NewTyp(lang.Int, false),
			}
			f.SetParent(right)
			right = f
			t = true

		case lang.Float64:
			f := &lang.FunctionCall{
				Name:   "float64",
				Args:   []lang.Expression{left},
				Return: lang.NewTyp(lang.Float64, false),
			}
			f.SetParent(left)
			left = f
			t = true
		}

	case lang.Float64:
		switch rt.String() {
		case lang.Bool:
			f := &lang.FunctionCall{
				Name:   "std.BoolToFloat64",
				Args:   []lang.Expression{right},
				Return: lang.NewTyp(lang.Float64, false),
			}
			f.SetParent(right)
			right = f
			t = true

		case lang.Int:
			f := &lang.FunctionCall{
				Name:   "float",
				Args:   []lang.Expression{right},
				Return: lang.NewTyp(lang.Float64, false),
			}
			f.SetParent(right)
			right = f
			t = true
		}
	}
	return t
}

// flowControlExpr parses next statement and tries
// to convert it into expression which can be used
// in the for, do, if. If it is expression as it should
// be, nothing happens, expression is returned.
// Extra work will be done with an assign and {inc,dec}rements.
// They will be moved to the "Init" section, condition will
// be replaced by a variable + possible convertion using std.Truthy.
// This is the first move from many, I want to resolve
// everything in the examples/33.php, but code is not ready
// for this yet.
func (p *parser) flowControlExpr(b lang.Block, n node.Node) (init lang.Node, expr lang.Expression) {
	s := p.simpleExpression(b, n)
	switch s.(type) {
	case *lang.Assign:
		a := s.(*lang.Assign)
		init = a
		expr = lang.NewVarRef(a.Left(), a.Left().CurrentType)

	case *lang.Dec:
		d := s.(*lang.Dec)
		init = d
		v := d.UsedVar()
		expr = lang.NewVarRef(v, v.CurrentType)

	case *lang.Inc:
		i := s.(*lang.Inc)
		init = i
		v := i.UsedVar()
		expr = lang.NewVarRef(v, v.CurrentType)

	case lang.Expression:
		expr = s.(lang.Expression)

	default:
		panic(`flowControlExpr: missing expression`)
	}
	if !expr.Type().Equal(lang.Bool) {
		var err error
		expr, err = p.funcs.Namespace("std").Call("Truthy", []lang.Expression{expr})
		if err != nil {
			panic(err)
		}
		expr.SetParent(b)
	}

	return
}

func (parser *parser) buildAssignment(parent lang.Block, name string, right lang.Expression) *lang.Assign {
	t := right.Type()
	if t.Equal(lang.Void) {
		panic("Cannot assign \"void\" " + "to \"" + name + "\".")
	}

	v := parent.HasVariable(name, false)
	fd := false
	if v == nil {
		v = lang.NewVariable(name, t, false)
		parent.DefineVariable(v)
		fd = true
	} else if v.CurrentType != t {
		if v.FirstDefinition.Parent() == parent {
			v.CurrentType = t
		} else {
			v = lang.NewVariable(name, t, false)
			fd = true
		}
		parent.DefineVariable(v)
	}

	as, err := lang.NewAssign(v, right)
	if err != nil {
		panic(err)
	}

	if fd {
		v.FirstDefinition = as
	}

	as.FirstDefinition = fd
	as.SetParent(parent)

	return as
}

// Function makes things much easier, I expect
// identifier name to be just simple right now
// defined string, no variable etc.
func (p *parser) identifierName(v *expr.Variable) string {
	switch v.VarName.(type) {
	case *node.Identifier:
		n := v.VarName.(*node.Identifier).Value
		return p.translator.Translate(n)

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
	return p.functionTranslator.Translate(s)
}

func (p *parser) servePrint(args []lang.Expression) (*lang.FunctionCall, error) {
	v := p.gc.HasVariable("W", false)
	if v == nil {
		return nil, errors.New("Variable io.Writer not defined")
	}
	args = append([]lang.Expression{lang.NewVarRef(v, lang.NewTyp(lang.Writer, false))}, args...)
	return p.funcs.Namespace("fmt").Call("Fprintf", args)
}

// TODO: Type could know this.
func varFormat(t string) string {
	switch t {
	case lang.Int:
		return "%d"

	case lang.Float64:
		return "%g"

	case lang.String:
		return "%s"

	default:
		return "%v"
	}
}
