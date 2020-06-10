package p

import (
	"errors"
	"fmt"

	"github.com/lSimul/php2go/lang"
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
		fn:        make(map[string][]*lang.Function),
	}

	fmt := map[string][]*lang.Function{
		"Print": {
			{
				Name: "Print",
				Args: []*lang.Variable{
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp(lang.Void, false),
			},
		},
		"Printf": {
			{
				Name: "Printf",
				Args: []*lang.Variable{
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp(lang.Void, false),
			},
		},
		"Fprintf": {
			{
				Name: "Fprintf",
				Args: []*lang.Variable{
					lang.NewVariable("W", lang.NewTyp(lang.Writer, false), false),
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp(lang.Void, false),
			},
		},
		"Sprintf": {
			{
				Name: "Sprintf",
				Args: []*lang.Variable{
					lang.NewVariable("format", lang.NewTyp(lang.String, false), false),
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp(lang.String, false),
			},
		},
	}
	fn.funcs["fmt"] = &funcs{
		namespace: "fmt",
		fn:        fmt,
	}

	std := map[string][]*lang.Function{
		"Concat": {
			{
				Name: "Concat",
				Args: []*lang.Variable{
					lang.NewVariable("left", lang.NewTyp(lang.Anything, false), false),
					lang.NewVariable("right", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.String, false),
			},
		},
		"StrDec": {
			{
				Name: "StrDec",
				Args: []*lang.Variable{
					lang.NewVariable("s", lang.NewTyp(lang.String, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.String, false),
			},
		},
		"StrInc": {
			{
				Name: "StrInc",
				Args: []*lang.Variable{
					lang.NewVariable("s", lang.NewTyp(lang.String, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.String, false),
			},
		},
		"Truthy": {
			{
				Name: "Truthy",
				Args: []*lang.Variable{
					lang.NewVariable("i", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.Bool, false),
			},
		},
		"ToInt": {
			{
				Name: "ToInt",
				Args: []*lang.Variable{
					lang.NewVariable("s", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.Int, false),
			},
		},
		"FileExists": {
			{
				Name: "FileExists",
				Args: []*lang.Variable{
					lang.NewVariable("s", lang.NewTyp(lang.String, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.Bool, false),
			},
		},
	}
	fn.funcs["std"] = &funcs{
		namespace: "github.com/lSimul/php2go/std",
		fn:        std,
	}

	arr := map[string][]*lang.Function{
		"NewScalar": {
			{
				Name: "NewScalar",
				Args: []*lang.Variable{
					lang.NewVariable("s", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp("array.Scalar", false),
			},
		},
		"NewInt": {
			{
				Name: "NewInt",
				Args: []*lang.Variable{
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp("array.Int", false),
			},
		},
		"NewString": {
			{
				Name: "NewString",
				Args: []*lang.Variable{
					lang.NewVariable("vals", lang.NewTyp(lang.Anything, false), false),
				},
				VariadicCount: true,

				Return: lang.NewTyp("array.String", false),
			},
		},
	}
	fn.funcs["array"] = &funcs{
		namespace: "github.com/lSimul/php2go/std/array",
		fn:        arr,
	}

	flag := map[string][]*lang.Function{
		"String": {
			{
				Name: "String",
				Args: []*lang.Variable{
					lang.NewVariable("name", lang.NewTyp(lang.String, false), false),
					lang.NewVariable("value", lang.NewTyp(lang.String, false), false),
					lang.NewVariable("usage", lang.NewTyp(lang.String, false), false),
				},
				VariadicCount: false,

				Return: lang.NewTyp(lang.String, true),
			},
		},
	}
	fn.funcs["flag"] = &funcs{
		namespace: "flag",
		fn:        flag,
	}

	return fn
}

type funcs struct {
	namespace string
	fn        map[string][]*lang.Function
}

type FileFunc struct {
	*Func

	file *lang.File
}

// Adds a function to the universe so it is recognized when calling
// the function using Func.Namespace(string).Call(string, []lang.Expression)
// First argument is a real name of the function. It goes well with the
// missingArgs argument, goal is to found the correct function. Functions
// in PHP can have variable amount of arguments.
func (fn *FileFunc) Add(name string, f *lang.Function, missingArgs int) {
	if _, ok := fn.funcs[""].fn[name]; !ok {
		fn.funcs[""].fn[name] = make([]*lang.Function, 0)
	}
	if len(fn.funcs[""].fn[name]) > missingArgs {
		fn.funcs[""].fn[name][missingArgs] = f
	} else {
		fn.funcs[""].fn[name] = append(fn.funcs[""].fn[name], f)
	}
	fn.file.Add(f)
}

func (f *FileFunc) Namespace(n string) *FunctionCaller {
	fn, ok := f.funcs[n]
	if !ok {
		panic(`Unknown namespace ` + n)
	}

	f.file.AddImport(fn.namespace)
	return &FunctionCaller{
		namespace: n,
		Func:      &fn.fn,
	}
}

type FunctionCaller struct {
	namespace string
	Func      *map[string][]*lang.Function
}

func (fc *FunctionCaller) NeedsGlobal(name string) {
	funcs, ok := (*fc.Func)[name]
	if !ok {
		return
	}
	for _, f := range funcs {
		f.NeedsGlobal = true
	}
}

func (fc *FunctionCaller) Call(name string, args []lang.Expression) (*lang.FunctionCall, error) {
	funcs, ok := (*fc.Func)[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Function '%s' is not defined.", name))
	}

	f := funcs[0]
	if !f.VariadicCount {
		i := len(f.Args) - len(args)
		if i >= len(funcs) {
			return nil, errors.New("Undefined function for this amount of parameters")
		}
		if i != 0 {
			f = funcs[i]
		}
	}

	// TODO: Refactor this.
	if f.VariadicCount {
		if len(f.Args)-1 > len(args) {
			return nil, errors.New("Not enough arguments.")
		}
		// Compare first N-1 args.
		for i := 0; i < len(f.Args)-1; i++ {
			// TODO: Implement better management of the types;
			// one if statement is not going to do it.
			if f.Args[i].Type().Equal(lang.Anything) {
				continue
			}
			if !args[i].Type().Eq(f.Args[i].Type()) {
				return nil, errors.New("Wrong argument type.")
			}
		}
		// Compare the rest, if any.
		ref := f.Args[len(f.Args)-1]
		for i := len(f.Args); i < len(args); i++ {
			if ref.Type().Equal(lang.Anything) {
				continue
			}
			if !args[i].Type().Eq(ref.Type()) {
				return nil, errors.New("Wrong argument type.")
			}
		}
	} else {
		if len(f.Args) != len(args) {
			return nil, errors.New("Wrong argument count.")
		}

		// TODO: Check for passing by reference.
		for i := 0; i < len(args); i++ {
			// TODO: Implement better management of the types;
			// one if statement is not going to do it.
			if f.Args[i].Type().Equal(lang.Anything) {
				continue
			}
			if t := f.Args[i].Type(); !args[i].Type().Eq(t) {
				if t.IsPointer && !args[i].Type().IsPointer {
					t, ok := args[i].(*lang.VarRef)
					if !ok {
						return nil, errors.New("Only variable can be send by reference.")
					}
					if err := t.ByReference(); err != nil {
						return nil, err
					}
				} else {
					return nil, errors.New("Wrong argument type.")
				}
			}
		}
	}

	n := ""
	if fc.namespace != "" {
		n = fc.namespace + "."
	}
	n += f.Name

	c := &lang.FunctionCall{
		Func: f,
		Name: n,

		Args:   args,
		Return: f.Return,
	}
	for _, a := range c.Args {
		a.SetParent(c)
	}
	return c, nil
}
