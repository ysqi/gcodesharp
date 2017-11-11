package gfmt

import (
	"go/build"
	"strings"
	"testing"

	"github.com/ysqi/gcodesharp/context"
)

func TestRun(t *testing.T) {
	ctx, err := context.New()
	if err != nil {
		t.Fatal(err)
	}

	p, err := build.Import("github.com/ysqi/gcodesharp/gfmt/testdata", "", build.IgnoreVendor)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Packages = append(ctx.Packages, p)
	s, err := New(ctx, func(service, msg string) {
		t.Fatalf("%s:%s", service, msg)
	})
	if err != nil {
		t.Fatal(err)
	}
	s.Run()
	s.Wait()
	if s.Report.GoFmt != gofmtpath {
		t.Fatal("need gofmt path value")
	}
	if len(s.Report.Files) == 2 {
		t.Fatal("need report two file")
	}
}

func TestGetGoFiles(t *testing.T) {
	files, err := getGoFiles(".")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) < 2 {
		t.Fatalf("want get more than two go file,but got %d items", len(files))
	}

	if !contains(files, "gfmt.go") {
		t.Fatal("want contain gfmt.go in files")
	}
	if !contains(files, "gfmt_test.go") {
		t.Fatal("want contain gfmt_test.go in files")
	}
}

func TestGoFmt(t *testing.T) {
	files, err := runGoFmt("gfmt_test.go")
	if err != nil {
		t.Fatal(err)
	}

	if !files[0].NeedFmt {
		t.Fatal("want need format but got no need")
	} else {
		diff := `-		&File{Name: "gfmt_test.go"}`
		if !strings.Contains(files[0].Diff, diff) {
			t.Fatal("need contains diff\n" + diff)
		}
	}

	files, err = runGoFmt("gfmt.go", "gfmt_test2.go")
	if err == nil {
		t.Fatal("need error,but got nil")
	}
	for _, f := range files {
		if f.NeedFmt {
			t.Fatal("need not set")
		}
	}

	files, err = runGoFmt("./testdata/needFmt.go", "./testdata/needFmt2.go")
	if err != nil {
		t.Fatal(err)
	}
	if !files[0].NeedFmt {
		t.Fatal("want need format,but got no need")
	}

	diff := `+	s := "hello"`
	if !strings.Contains(files[0].Diff, diff) {
		t.Fatal("need contains diff\n" + diff)

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
