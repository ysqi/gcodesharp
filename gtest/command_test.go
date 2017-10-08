package gtest

import (
	"strings"
	"testing"

	"github.com/ysqi/gcodesharp/context"
)

func TestPackageList(t *testing.T) {
	list, err := getPackageList("github.com/ysqi/gcodesharp")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(list, "github.com/ysqi/gcodesharp") {
		t.Fatalf("go list %+v need contains %q", list, "github.com/ysqi/gcodesharp")
	}
	if !contains(list, "github.com/ysqi/gcodesharp/gtest") {
		t.Fatalf("go list %+v need contains %q", list, "github.com/ysqi/gcodesharp/gtest")
	}

	list1, err := getPackageList("")
	if err != nil {
		t.Fatal(err)
	}
	list2, err := getPackageList(".")
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

func TestRelTime(t *testing.T) {
	ctx, err := context.New()
	if err != nil {
		t.Fatal(err)
	}
	report, err := Run(ctx, &Config{
		PackageDirs: []string{
			"./testdata",
		},
		ContainImport: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Packages) != 1 {
		t.Fatal("want one pakcage rest report,but zero")
	}
	pkg := report.Packages[0]
	for _, u := range pkg.Units {
		t.Logf("%+v", u)
	}
	if !pkg.Failed {
		t.Fatal("want test failed, but pass")
	}
	// if pkg.Err != "exit status 1" {
	// 	t.Fatalf("want error 'exit status 1' got '%s'", pkg.Err)
	// }
	if pkg.Cost == 0.0 {
		t.Fatal("want get package run test speed time")
	}
	name := "github.com/ysqi/gcodesharp/gtest/testdata"
	if pkg.Name != name {
		t.Fatalf("want get package name %q, got %q", name, pkg.Name)
	}
	pass, skip, fail := pkg.PassCount(), pkg.SkipCount(), pkg.FailCount()
	if pass != 2 {
		t.Fatalf("want the number of pass test is %d,but got %d", 2, pass)
	}
	if fail != 3 {
		t.Fatalf("want the number of fail test is %d,but got %d", 3, fail)
		s := pkg.GetByResult(FAIL)
		last := s[len(s)-1]
		want := "panic: 3.this is a panic test info [recovered]"
		if !strings.HasPrefix(last.Output, want) {
			t.Fatalf("want the panic test output muse be start with %q,but got %q", want, last.Output)
		}
	}
	if skip != 1 {
		t.Fatalf("want the number of skip test is %d,but got %d", 1, skip)
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
