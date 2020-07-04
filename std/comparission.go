package std

// Truthy converts anything to boolean.
// Used to simulate implicit conversion.
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

// Bool is used to add an option
// to implement non-implicit conversion.
// Used in std.Array and std.SQL.
type Bool interface {
	// ToBool is the only method implemented
	// in this interface. The name is specific
	// enough so no collisions should be ever
	// discovered.
	ToBool() bool
}
