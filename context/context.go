package context

import (
	"errors"
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
