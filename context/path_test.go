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

package context

import "testing"

func TestPackageList(t *testing.T) {
	list, err := GetPackagePaths("github.com/ysqi/gcodesharp/...")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(list, "github.com/ysqi/gcodesharp") {
		t.Fatalf("go list %+v need contains %q", list, "github.com/ysqi/gcodesharp")
	}
	if !contains(list, "github.com/ysqi/gcodesharp/gtest") {
		t.Fatalf("go list %+v need contains %q", list, "github.com/ysqi/gcodesharp/gtest")
	}

	list1, err := GetPackagePaths("")
	if err != nil {
		t.Fatal(err)
	}
	list2, err := GetPackagePaths(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(list1) != len(list2) {
		t.Fatal("get diffrent result")
	}
	for i := 0; i < len(list1); i++ {
		if list1[i] != list2[i] {
			t.Fatal("get diffrent result")
		}
	}

}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
