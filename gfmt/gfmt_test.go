package gfmt

import (
	"strings"
	"testing"

	"github.com/ysqi/gcodesharp/context"
)

func TestRun(t *testing.T) {
	ctx, err := context.New()
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		PackagePaths: []string{"github.com/ysqi/gcodesharp"},
	}
	report, err := Run(ctx, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if report.GoFmt != gofmtpath {
		t.Fatal("need gofmt path value")
	}
	if len(report.Files) == 0 {
		t.Fatal("need more one go file to check go format")
	}
	for _, f := range report.Files {
		if f.NeedFmt && f.Name != "gfmt_test.go" {
			t.Fatalf("%q want no need gofmt but got need", f.Name)
		}
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
	files := []*File{
		&File{Name: "gfmt_test.go"},
	}
	err := runGoFmt(files)
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

	files = []*File{
		&File{Name: "gfmt.go"},
		&File{Name: "gfmt_test2.go"},
	}
	if err = runGoFmt(files); err == nil {
		t.Fatal("need error,but got nil")
	}
	for _, f := range files {
		if f.NeedFmt {
			t.Fatal("need not set")
		}
	}

	files = []*File{
		&File{Name: "./testdata/needFmt.go"},
		&File{Name: "./testdata/needFmt2.go"},
	}
	if err = runGoFmt(files); err != nil {
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
