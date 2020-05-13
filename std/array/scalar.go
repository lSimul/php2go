package array

import (
	"math"
	"strconv"
)

// TODO: PHP do not think that it is always a string,
// but it can be overriden ignoring the type:
// 0 == "0" == 0.1 == false
// Part of this issue is resolved using IntValue().

// Scalar well-defines a key for PHP associative array.
// It is a decorated string which can convert other
// types to string using PHP common conventions, like
// false => zero, NULL => "" and so on.
type Scalar string

// IntValue translates value of the scalar to int,
// if possible.
// This is used to find out what the next index in
// the array will be, if values are added using Add/Push.
func (s Scalar) IntValue() (int, bool) {
	v, err := strconv.Atoi(string(s))
	if err != nil {
		return v, false
	}
	return v, true
}

// NewScalar converts primitive data types and
// NULL to strng, which can be used to index
// a PHP array.
func NewScalar(val interface{}) Scalar {
	// TODO: Pointers can appear here.
	switch t := val.(type) {
	case bool:
		if t {
			return Scalar("1")
		}
		return Scalar("0")

	case int:
		return Scalar(strconv.Itoa(t))

	case float64:
		i := int(math.Floor(t))
		return Scalar(strconv.Itoa(i))

	case string:
		return Scalar(t)
	}
	if val != nil {
		panic(`Only NULL and primitive types can be an array key.`)
	}
	return ""
}

// Array encapsulates common methods
// for arrays. That means no Add/Edit/Push/Iter,
// because these are strongly typed. Only the ones
// using only scalar as an argument are here.
type Array interface {
	Isset(Scalar) bool
	Unset(Scalar)
	Count() int
}
