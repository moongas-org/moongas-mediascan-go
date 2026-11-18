package utils

import "testing"

// Check with a nil error must be a no-op. The non-nil path exits the process,
// which exits the process and so cannot be exercised in a unit test.
func TestCheckNilError(t *testing.T) {
	Check(nil, "must not exit")
}
