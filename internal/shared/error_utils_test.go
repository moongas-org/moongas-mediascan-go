package shared

import "testing"

// Check with a nil error must be a no-op. The non-nil path calls log.Fatalf,
// which exits the process and so cannot be exercised in a unit test.
func TestCheckNilError(t *testing.T) {
	Check(nil, "must not exit")
}
