package shared

import "strings"

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// ContainsAnyOf returns true if the given string (a) contains any of the
// substrings in the given slice of strings (list) (case-insensitive).
func ContainsAnyOf(a string, list []string) bool {
	if len(a) > 0 {
		a = strings.ToLower(a)
		for _, b := range list {
			if strings.Contains(a, strings.ToLower(b)) {
				return true
			}
		}
	}
	return false
}
