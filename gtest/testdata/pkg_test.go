package testdata

import "testing"

func TestPass1(t *testing.T) {
	t.Log("1.this is check pass info")
}
func TestPass2(t *testing.T) {
	t.Log("2.this is check pass info")
}
func TestFail1(t *testing.T) {
	t.Fatal("1.this is check failed info")
}
func TestFail2(t *testing.T) {
	t.Fatal("2.this is check failed info")
}

func TestSkip(t *testing.T) {
	t.Skip("1.this is a skip test")
}

// func TestTimeout(t *testing.T) {
// 	time.Sleep(1 * time.Minute)
// }

// must write at last
func TestPainc(t *testing.T) {
	panic("3.this is a panic test info")
}
