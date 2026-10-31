package utils

// Pointer returns a pointer to a copy of value.
func Pointer[T any](value T) *T {
	return &value
}
