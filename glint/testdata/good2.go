// Test

package testdata

func g2(f func() bool) string {
	if ok := f(); ok {
		return "it's okay"
	}
	return "it's NOT okay!"
}
