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
	s, err := New(ctx, func(fmt_ string, args ...interface{}) {
		t.Fatalf(fmt_, args...)
	})
	if err != nil {
		t.Fatal(err)
	}
	s.Run()
	s.Wait()
	if s.Report.GoFmt != gofmtpath {
		t.Fatal("need gofmt path value")
	}
	if len(s.Report.Files) != 2 {
		t.Fatalf("need report two file ,got %d files", len(s.Report.Files))
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

	if files[0].NeedFmt {
		t.Fatal("want need format but got no need")
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
