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

import (
	"errors"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Context represents the current project context.
type Context struct {
	GopathList []string // List of GOPATHs in environment. Includes "src" dir.
	Goroot     string   // The path to the standard library.
	GoEnv      map[string]string

	// Packages is list of need handle package
	Packages []*build.Package
}

// New create a new context.
func New() (*Context, error) {
	env, err := getGoEnv()
	if err != nil {
		return nil, err
	}
	goroot := env["GOROOT"]
	if goroot == "" {
		return nil, errors.New("Unable to determine GOROOT")
	}
	goroot = filepath.Join(goroot, "src")
	if _, err := os.Stat(goroot); err != nil {
		return nil, err
	}
	all := env["GOPATH"]
	// Get the GOPATHs. Prepend the GOROOT to the list.
	if len(all) == 0 {
		return nil, errors.New("Missing GOPATH. Check your environment variable GOPATH")
	}
	gopathList := filepath.SplitList(all)
	gopathGoroot := make([]string, 0, len(gopathList)+1)
	gopathGoroot = append(gopathGoroot, goroot)
	for _, gopath := range gopathList {
		srcPath := filepath.Join(gopath, "src") + string(filepath.Separator)
		srcPathEvaled, err := filepath.EvalSymlinks(srcPath)
		if err != nil {
			return nil, err
		}
		gopathGoroot = append(gopathGoroot, srcPath, srcPathEvaled+string(filepath.Separator))
	}

	ctx := &Context{
		GopathList: gopathGoroot,
		Goroot:     goroot,
		GoEnv:      env,
	}
	return ctx, nil
}

func getGoEnv() (map[string]string, error) {
	env := map[string]string{}
	cmd := exec.Command("go", "env")
	var goEnv []byte
	goEnv, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(goEnv), "\n") {
		if k, v, ok := parseGoEnvLine(line); ok {
			env[k] = v
		}
	}
	return env, nil
}
