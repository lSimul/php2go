package p

import (
	"errors"
	"php2go/lang"
)

type Func struct {
	funcs map[string]*funcs

	gc *lang.GlobalContext
}

func NewFunc(gc *lang.GlobalContext) *Func {
	fn := &Func{
		funcs: make(map[string]*funcs),
		gc:    gc,
	}

	fn.funcs[""] = &funcs{
		namespace: "",
		fn:        &gc.Funcs,
		used:      true,
	}

	fmt := make(map[string]*lang.Function)
	fn.funcs["fmt"] = &funcs{
		namespace: "fmt",
		fn:        &fmt,
		used:      false,
	}

	std := map[string]*lang.Function{
		"Concat": {
			Name: "Concat",
			Args: []*lang.Variable{
				lang.NewVariable("left", lang.Anything, false),
				lang.NewVariable("right", lang.Anything, false),
			},
			VariadicCount: false,

			Return: lang.String,
		},
	}
	fn.funcs["std"] = &funcs{
		namespace: "php2go/std",
		fn:        &std,
		used:      false,
	}

	arr := map[string]*lang.Function{
		"NewScalar": {
			Name: "NewScalar",
			Args: []*lang.Variable{
				lang.NewVariable("s", lang.Anything, false),
			},
			VariadicCount: false,

			Return: "array.Scalar",
		},
	}
	fn.funcs["array"] = &funcs{
		namespace: "php2go/std/array",
		fn:        &arr,
		used:      false,
	}

	return fn
}

type funcs struct {
	namespace string
	fn        *map[string]*lang.Function
	used      bool
}

func (fn *Func) Add(f *lang.Function) {
	fn.gc.Add(f)
}

func (f *Func) Namespace(n string) *FunctionCaller {
	fn, ok := f.funcs[n]
	if !ok {
		panic(`Unknown namespace ` + n)
	}
	if !fn.used {
		fn.used = true
		f.gc.AddImport(fn.namespace)
	}

	return &FunctionCaller{
		namespace: n,
		functions: fn.fn,
	}
}

type FunctionCaller struct {
	namespace string
	functions *map[string]*lang.Function
}

func (fc *FunctionCaller) Call(name string, args []lang.Expression) (*lang.FunctionCall, error) {
	f, ok := (*fc.functions)[name]
	if !ok {
		return nil, errors.New("Function is not defined.")
	}

	if len(f.Args) != len(args) {
		return nil, errors.New("Wrong argument count.")
	}

	// TODO: Check for passing by reference.
	for i := 0; i < len(args); i++ {
		// TODO: Implement better management of the types;
		// one if statement is not going to do it.
		if f.Args[i].Type() == lang.Anything {
			continue
		}
		if args[i].Type() != f.Args[i].Type() {
			return nil, errors.New("Wrong argument type.")
		}
	}

	n := ""
	if fc.namespace != "" {
		n = fc.namespace + "."
	}
	n += f.Name

	c := &lang.FunctionCall{
		Name:   n,
		Args:   args,
		Return: f.Return,
	}
	for _, a := range c.Args {
		a.SetParent(c)
	}
	return c, nil
}
