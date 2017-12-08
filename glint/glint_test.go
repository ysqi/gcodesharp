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

package glint

import (
	"go/build"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ysqi/gcodesharp/context"
)

func TestRun(t *testing.T) {
	ctx, err := context.New()
	if err != nil {
		t.Fatal(err)
	}

	p, err := build.Import("github.com/ysqi/gcodesharp/glint/testdata", "", build.IgnoreVendor)
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

	if f := find(t, s.Report.Files, "bad1.go"); f == nil {
		t.Fatal("want report ba1.go problem")
	} else if con := f.ProblemContent(); !strings.Contains(con, "Func") {
		t.Fatalf("want report a func problem in ba1.go, got %s", con)
	} else if !strings.Contains(con, "myZeroInt") {
		t.Fatalf("want report var myZeroInt in ba1.go, got %s", con)
	}

	if f := find(t, s.Report.Files, "bad2.go"); f == nil {
		t.Fatal("want report ba1.go problem")
	} else if con := f.ProblemContent(); !strings.Contains(con, "if block") {
		t.Fatalf("want report if block problem in ba1.go, got %s", con)
	}

	if f := find(t, s.Report.Files, "good1.go"); f.HasProblem() {
		t.Fatalf("want no problem in good.go, got %s", f.ProblemContent())
	}
	if f := find(t, s.Report.Files, "good2.go"); f.HasProblem() {
		t.Fatalf("want no problem in good2.go, got %s", f.ProblemContent())
	}
}

func find(t *testing.T, l []*File, name string) *File {
	for _, f := range l {
		if filepath.Base(f.Name) == name {
			return f
		}
	}
	t.Fatalf("not find file %s", name)
	return nil
}
