package lang

import "testing"

func TestLang(t *testing.T) {
	t.Run("Constructions", constructors)
}

func constructors(t *testing.T) {
	gc := NewGlobalContext()
	if gc.Vars == nil {
		t.Error("Context does not have initialized Vars.")
	}
	if len(gc.Vars) != 0 {
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
		Name: "",
		Type: String,
	}
	_, err = NewAssign(v, nil)
	if err == nil {
		t.Error("Missing right side.")
	}

	_, err = NewAssign(nil, v)
	if err == nil {
		t.Error("Missing left side.")
	}

	a, err = NewAssign(v, v)
	if a.GetType() != v.GetType() {
		t.Errorf("'%s' expected, '%s' found.\n", v.GetType(), a.GetType())
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
	_, err = NewBinaryOp("+", v, nil)
	if err == nil {
		t.Error("Right expression is missing.")
	}
	_, err = NewBinaryOp("+", nil, v)
	if err == nil {
		t.Error("Left expression is missing.")
	}
	bo, _ := NewBinaryOp("+", v, v)
	if bo.Left != v {
		t.Error("Wrong left side.")
	}
	if bo.Right != v {
		t.Error("Wrong right side.")
	}
	if bo.inBrackets {
		t.Error("No brackets needed.")
	}

	void := &Variable{
		Type: Void,
	}
	_, err = NewBinaryOp("", void, v)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
	_, err = NewBinaryOp("", void, void)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
	_, err = NewBinaryOp("", v, void)
	if err == nil {
		t.Error("Binary op cannot be used with 'void'")
	}
}
