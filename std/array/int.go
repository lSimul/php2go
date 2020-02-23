package array

var _ Array = (*Int)(nil)

type Int struct {
	associative map[Scalar]int
	order       []int
	lastIndex   int
}

func NewInt(vals ...int) Int {
	a := Int{
		associative: make(map[Scalar]int),
		order:       make([]int, 0),
		lastIndex:   0,
	}
	a.Add(vals...)
	return a
}

func (a *Int) Add(vals ...int) *Int {
	for _, v := range vals {
		k := NewScalar(a.lastIndex)
		a.add(k, v)
		a.lastIndex++
	}
	return a
}

func (a *Int) Push(vals ...int) int {
	a.Add(vals...)
	return len(a.order)
}

func (a *Int) Edit(k Scalar, v int) *Int {
	if i, ok := a.associative[k]; ok {
		a.order[i] = v
	} else if i, ok := k.IntValue(); ok && i > a.lastIndex {
		a.lastIndex = i
		a.Add(v)
	} else {
		a.add(k, v)
	}
	return a
}

func (a *Int) add(k Scalar, v int) {
	a.order = append(a.order, v)
	a.associative[k] = len(a.order) - 1
}

func (a Int) At(k Scalar) int {
	if v, ok := a.associative[k]; ok {
		return v
	}
	panic("undefined index " + k)
}

func (a Int) Iter() []int {
	return a.order
}

func (a Int) Isset(k Scalar) bool {
	_, ok := a.associative[k]
	return ok
}

func (a *Int) Unset(k Scalar) {
	i, ok := a.associative[k]
	if !ok {
		return
	}
	delete(a.associative, k)

	copy(a.order[i:], a.order[i+1:])
	a.order = a.order[:len(a.order)-1]
	for k, v := range a.associative {
		if v > i {
			a.associative[k] = v - 1
		}
	}
}
