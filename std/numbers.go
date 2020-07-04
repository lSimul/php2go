package std

import (
	"regexp"
	"strconv"
)

// ToInt transfers anything to int.
// The main part is conversion from
// string to int. It does the same thing
// as (int) in PHP, that means it tries to
// parse it as long it looks like an int.
// BUG(ls): it does not support diferent
// bases like HEX etc.
func ToInt(s interface{}) int {
	switch s := s.(type) {
	case int:
		return s

	case bool:
		if s {
			return 1
		}
		return 0

	case string:
		if len(s) == 0 {
			return 0
		}
		r, _ := regexp.Compile("[0-9]+")
		s = r.FindString(s)
		res, _ := strconv.Atoi(s)
		return res

	case float32:
		return int(s)
	case float64:
		return int(s)
	}
	if s == nil {
		return 0
	}
	return 1
}
