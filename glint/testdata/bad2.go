// Test

package testdata

func g(f func() bool) string {
	if ok := f(); ok {
		return "it's okay"
	} else { // MATCH
		return "it's NOT okay!"
	}
}
