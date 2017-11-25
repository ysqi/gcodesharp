// Copyright (C) 2017. author ysqi(devysq@gmail.com).
//
// The gcodesharp is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gcodesharp is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
