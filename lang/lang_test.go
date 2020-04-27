package lang

import "testing"

func TestLang(t *testing.T) {
	t.Run("Constructions", constructors)
}

func constructors(t *testing.T) {
	gc := NewGlobalContext()
	if gc.vars == nil {
		t.Error("Context does not have initialized Vars.")
	}
	if len(gc.vars) != 0 {
		t.Error("Context has some extra vars.")
	}

	if gc.Funcs == nil {
		t.Error("Context does not have initialized Funcs.")
	}
	if len(gc.Funcs) != 0 {
		t.Error("Context has some extra functions.")
	}

	var parent Block = nil
	c := NewCode(parent)
	if c.Parent() != parent {
		t.Error("Wrong parent set.")
	}
	if c.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(c.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if c.Statements == nil {
		t.Error("Code does not have initialized Statements.")
	}
	if len(c.Statements) != 0 {
		t.Error("Code has some extra statements.")
	}

	parent = &Code{}
	c = NewCode(parent)
	if c.Parent() != parent {
		t.Error("Wrong parent set.")
	}
	if c.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(c.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if c.Statements == nil {
		t.Error("Code does not have initialized Statements.")
	}
	if len(c.Statements) != 0 {
		t.Error("Code has some extra statements.")
	}

	parent = nil
	f := ConstructFor(parent)
	if f.Parent() != parent {
		t.Error("Wrong parent set.")
	}
	if f.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(f.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if f.Block.Parent() != f {
		t.Error("Block does not have set up parent.")
	}
	if f.Block.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(f.Block.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if f.Block.Statements == nil {
		t.Error("Code does not have initialized Statements.")
	}
	if len(f.Block.Statements) != 0 {
		t.Error("Code has some extra statements.")
	}

	parent = &Code{}
	f = ConstructFor(parent)
	if f.Parent() != parent {
		t.Error("Wrong parent set.")
	}
	if f.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(f.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if f.Block.Parent() != f {
		t.Error("Block does not have set up parent.")
	}
	if f.Block.Vars == nil {
		t.Error("Code does not have initialized Vars.")
	}
	if len(f.Block.Vars) != 0 {
		t.Error("Code has some extra vars.")
	}
	if f.Block.Statements == nil {
		t.Error("Code does not have initialized Statements.")
	}
	if len(f.Block.Statements) != 0 {
		t.Error("Code has some extra statements.")
	}

	var a *Assign
	_, err := NewAssign(nil, nil)
	if err == nil {
		t.Error("Nothing can be created from nils.")
	}

	v := &Variable{
		Name:        "",
		typ:         String,
		CurrentType: String,
	}
	vr := &VarRef{
		V:   v,
		typ: String,
	}

	_, err = NewAssign(v, nil)
	if err == nil {
		t.Error("Missing right side.")
	}

	_, err = NewAssign(nil, vr)
	if err == nil {
		t.Error("Missing left side.")
	}

	a, err = NewAssign(v, vr)
	if a.Type() != v.Type() {
		t.Errorf("'%s' expected, '%s' found.\n", v.Type(), a.Type())
	}
	if a.FirstDefinition {
		t.Error("Default value is false, not true.")
	}
	// TODO: Add option to test left side.
	// Check more than that is it set, right
	// side cannot be void and more.
	if a.left != v {
		t.Error("Wrong left side.")
	}

	_, err = NewBinaryOp("+", nil, nil)
	if err == nil {
		t.Error("Left expression is missing.")
	}
	_, err = NewBinaryOp("+", vr, nil)
	if err == nil {
		t.Error("Right expression is missing.")
	}
	_, err = NewBinaryOp("+", nil, vr)
	if err == nil {
		t.Error("Left expression is missing.")
	}
	bo, _ := NewBinaryOp("+", vr, vr)
	if bo.left != vr {
		t.Error("Wrong left side.")
	}
	if bo.right != vr {
		t.Error("Wrong right side.")
	}
	if bo.inBrackets {
		t.Error("No brackets needed.")
	}

	void := &VarRef{
		V: &Variable{
			typ:         Void,
			CurrentType: Void,
		},
		typ: Void,
	}

	_, err = NewBinaryOp("", void, vr)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
	_, err = NewBinaryOp("", void, void)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
	_, err = NewBinaryOp("", vr, void)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
}
