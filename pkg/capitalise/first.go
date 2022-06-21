package capitalise

import (
	"unicode"
)

// Capitalises the first character of a string.
func First(str string) string {
	if str == "" {
		return ""
	}

	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])

	return string(tmp)
}
