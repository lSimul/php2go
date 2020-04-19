package p

import (
	"bytes"
	"fmt"
)

const (
	Public  = 1
	Private = 2
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
}

type NameTranslation interface {
	Translate(string, int) string
}

type nameTranslator struct {
	names map[string]string
	used  map[string]bool
}

func (t *nameTranslator) Translate(name string, visibility int) string {
	switch visibility {
	case Public:
		b := []byte(name)
		b[0] = bytes.ToUpper(b)[0]
		name = string(b)
	case Private:
		b := []byte(name)
		b[0] = bytes.ToLower(b)[0]
		name = string(b)
	}

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
	if _, used := t.used[n]; used {
		t.resolveConflict(name, try+1)
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

func NewFunctionTranslator() NameTranslation {
	used := make(map[string]bool)
	for k, v := range keywords {
		used[k] = v
	}
	used["main"] = true
	return &nameTranslator{
		names: make(map[string]string),
		used:  used,
	}
}
