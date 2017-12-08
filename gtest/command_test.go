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

package gtest

import (
	"go/build"
	"strings"
	"testing"

	"github.com/ysqi/gcodesharp/context"
)

func TestRelTime(t *testing.T) {
	ctx, err := context.New()
	if err != nil {
		t.Fatal(err)
	}

	p, err := build.Import("github.com/ysqi/gcodesharp/gtest/testdata", "", build.IgnoreVendor)
	if err != nil {
		t.Fatal(err)
	}
	ctx.Packages = append(ctx.Packages, p)

	ser, err := New(ctx, func(fm string, args ...interface{}) {
		t.Fatalf(fm, args...)
	})
	if err != nil {
		t.Fatal(err)
	}
	err = ser.Run()
	if err != nil {
		t.Fatal(err)
	}
	err = ser.Wait()
	if err != nil {
		t.Fatal(err)
	}

	if len(ser.Report.Packages) != 1 {
		t.Fatal("want one pakcage rest report,but zero")
	}
	pkg := ser.Report.Packages[0]
	for _, u := range pkg.Units {
		t.Logf("%+v", u)
	}
	if !pkg.Failed {
		t.Fatal("want test failed, but pass")
	}
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
