package p

import (
	"bytes"
	"fmt"
	"strings"
)

var keywords = map[string]bool{
	"break":       true,
	"case":        true,
	"chan":        true,
	"const":       true,
	"continue":    true,
	"default":     true,
	"defer":       true,
	"else":        true,
	"fallthrough": true,
	"for":         true,
	"func":        true,
	"go":          true,
	"goto":        true,
	"if":          true,
	"import":      true,
	"interface":   true,
	"map":         true,
	"package":     true,
	"range":       true,
	"return":      true,
	"select":      true,
	"struct":      true,
	"switch":      true,
	"type":        true,
	"var":         true,

	"array": true,
	"std":   true,

	// For globals implementation
	"g":      true,
	"global": true,
}

// ArrayType formats array type name
// for given type.
// The convention is simple, it is package
// name "array" and exported s.
func ArrayType(s string) string {
	return "array." + FirstUpper(s)
}

// ArrayItem is a reverse function
// to ArrayType, formating type
// present in the array.
func ArrayItem(s string) string {
	s = strings.TrimLeft(s, "array.")
	return FirstLower(s)
}

// IsArray checks, if the s is a name
// suitable for array type.
func IsArray(s string) bool {
	return strings.HasPrefix(s, "array.")
}

// FirstUpper uppercases first letter
// in s.
func FirstUpper(s string) string {
	b := []byte(s)
	b[0] = bytes.ToUpper(b)[0]
	return string(b)
}

// FirstLower lowercases first letter
// in s.
func FirstLower(s string) string {
	b := []byte(s)
	b[0] = bytes.ToLower(b)[0]
	return string(b)
}

type NameTranslation interface {
	Translate(string) string
}

type nameTranslator struct {
	names map[string]string
	used  map[string]bool
}

func (t *nameTranslator) Translate(name string) string {
	if name, defined := t.names[name]; defined {
		return name
	}

	new := name
	if _, used := t.used[name]; used {
		new = t.resolveConflict(name, 1)
	}
	t.names[name] = new
	t.used[new] = true
	return new
}

func (t nameTranslator) resolveConflict(name string, try int) string {
	n := fmt.Sprintf("%s%d", name, try)
	if used := t.used[n]; used {
		return t.resolveConflict(name, try+1)
	}
	return n
}

func NewNameTranslator() NameTranslation {
	used := make(map[string]bool)
	for k, v := range keywords {
		used[k] = v
	}
	return &nameTranslator{
		names: make(map[string]string),
		used:  used,
	}
}

type fnTranslator struct {
	nameTranslator
}

func (f *fnTranslator) Translate(name string) string {
	if name == "mainFunc" {
		name = f.nameTranslator.resolveConflict(name, 0)
	} else {
		name = strings.ToLower(name)
	}
	return f.nameTranslator.Translate(name)
}

func NewFunctionTranslator() NameTranslation {
	used := make(map[string]bool)
	for k, v := range keywords {
		used[k] = v
	}
	used["main"] = true
	used["mainServer"] = true
	used["mainCLI"] = true
	return &fnTranslator{
		nameTranslator{
			names: make(map[string]string),
			used:  used,
		},
	}
}

type labelTranslator struct {
	nameTranslator
}

func NewLabelTranslator() *labelTranslator {
	return &labelTranslator{
		nameTranslator: nameTranslator{
			names: make(map[string]string),
			used:  make(map[string]bool),
		},
	}
}

func (l *labelTranslator) Label(name string, unique bool, number int) string {
	if unique {
		new := l.resolveConflict(name, number)
		l.names[name] = new
		l.used[new] = true
		return new
	}
	return l.Translate(name)
}
