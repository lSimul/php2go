package std

import (
	"regexp"
	"strconv"
)

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
