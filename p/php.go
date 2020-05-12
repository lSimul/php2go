package p

import (
	"errors"

	"php2go/lang"
)

var PHPFunctions = map[string](func(lang.Block, []lang.Expression) (*lang.FunctionCall, error)){
	"array_push": arrayPush,
	// "echo":       true, // extra case, AST does not use echo as a function
}

func arrayPush(b lang.Block, args []lang.Expression) (*lang.FunctionCall, error) {
	if len(args) < 2 {
		return nil, errors.New("array_push requires atlast two arguments")
	}

	v, ok := args[0].(*lang.VarRef)
	if !ok || !IsArray(v.Type().String()) {
		return nil, errors.New("First argument has to be a variable, an array.")
	}
	typ := ArrayItem(v.Type().String())

	vars := []lang.Expression{}
	for _, arg := range args[1:] {
		if !arg.Type().Equal(typ) {
			return nil, errors.New("Cannot push this type.")
		}
		vars = append(vars, arg)
	}

	fc := &lang.FunctionCall{
		Name:   v.V.Name + ".Push",
		Args:   vars,
		Return: lang.NewTyp(lang.Int, false),
	}

	fc.SetParent(b)
	return fc, nil
}
