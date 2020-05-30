package std

// TODO: Add arrays and something like
// comparable interface.
func Truthy(i interface{}) bool {
	switch v := i.(type) {
	case int:
		return v != 0

	case float64:
		return v != 0

	case string:
		return v != ""

	case bool:
		// Is this even needed?
		return v

	case Bool:
		return v.ToBool()
	}

	return i != nil
}

type Bool interface {
	ToBool() bool
}
