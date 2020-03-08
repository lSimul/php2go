package std

// TODO: Add arrays and something like
// comparable interface.
func Truthy(i interface{}) bool {
	switch i.(type) {
	case int:
		return i.(int) != 0

	case float64:
		return i.(float64) != 0

	case string:
		return i.(string) != ""

	case bool:
		// Is this even needed?
		return i.(bool)
	}

	return i != nil
}
