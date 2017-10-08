package context

import "testing"

func TestPackageList(t *testing.T) {
	list, err := GetPackagePaths("github.com/ysqi/gcodesharp")
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
