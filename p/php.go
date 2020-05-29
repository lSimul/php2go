package p

import (
	"errors"

	"github.com/lSimul/php2go/lang"
)

var PHPFunctions = map[string](func(lang.Block, []lang.Expression) (*lang.FunctionCall, error)){
	"array_push": arrayPush,
	"count":      count,

	"mysqli_connect":   mysqliConnect,
	"mysqli_select_db": mysqliSelectDB,
	"mysqli_query":     mysqliQuery,
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

func count(b lang.Block, args []lang.Expression) (*lang.FunctionCall, error) {
	if len(args) != 1 {
		return nil, errors.New("count requires exactly one argument")
	}

	v, ok := args[0].(*lang.VarRef)
	if !ok || !IsArray(v.Type().String()) {
		return nil, errors.New("Argument has to be a variable, an array.")
	}

	fc := &lang.FunctionCall{
		Name:   v.V.Name + ".Count",
		Return: lang.NewTyp(lang.Int, false),
	}

	fc.SetParent(b)
	return fc, nil
}

func mysqliConnect(b lang.Block, args []lang.Expression) (*lang.FunctionCall, error) {
	if len(args) != 3 {
		return nil, errors.New("mysqli_connect has to have three arguments.")
	}

	fc := &lang.FunctionCall{
		Name:   "std.NewSQL",
		Args:   args,
		Return: lang.NewTyp(lang.SQL, true),
	}

	fc.SetParent(b)
	return fc, nil
}

func mysqliSelectDB(b lang.Block, args []lang.Expression) (*lang.FunctionCall, error) {
	if len(args) != 2 {
		return nil, errors.New("mysqli_select_db requires exactly two arguments.")
	}

	v, ok := args[0].(*lang.VarRef)
	if !ok {
		return nil, errors.New("First argument should be a varref.")
	}
	if v.Type() != lang.NewTyp(lang.SQL, true) {
		return nil, errors.New("First argument is not of a type std.SQL.")
	}

	if args[1].Type() != lang.NewTyp(lang.String, false) {
		return nil, errors.New("Database name has to be a string.")
	}

	fc := &lang.FunctionCall{
		Name:   v.V.Name + ".SelectDB",
		Args:   args[1:],
		Return: lang.NewTyp(lang.SQL, false),
	}

	fc.SetParent(b)
	return fc, nil
}

func mysqliQuery(b lang.Block, args []lang.Expression) (*lang.FunctionCall, error) {
	if len(args) != 2 {
		return nil, errors.New("mysqli_query requires exactly two arguments.")
	}

	v, ok := args[0].(*lang.VarRef)
	if !ok {
		return nil, errors.New("First argument should be a varref.")
	}
	if v.Type() != lang.NewTyp(lang.SQL, true) {
		return nil, errors.New("First argument is not of a type std.SQL.")
	}

	if args[1].Type() != lang.NewTyp(lang.String, false) {
		return nil, errors.New("Query has to be a string.")
	}

	fc := &lang.FunctionCall{
		Name:   v.V.Name + ".Query",
		Args:   args[1:],
		Return: lang.NewTyp(lang.SQL, false),
	}

	fc.SetParent(b)
	return fc, nil
}
