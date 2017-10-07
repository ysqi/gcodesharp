package context

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// FindImportPath takes a absolute directory and returns the import path and go path.
func (ctx *Context) FindImportPath(dir string) (importPath, gopath string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}
	dirResolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", "", err
	}
	dirs := make([]string, 1)
	dirs = append(dirs, dir)
	if dir != dirResolved {
		dirs = append(dirs, dirResolved)
	}

	for _, gopath := range ctx.GopathList {
		for _, dir := range dirs {
			if fileHasPrefix(dir, gopath) || fileStringEquals(dir, gopath) {
				importPath = fileTrimPrefix(dir, gopath)
				importPath = slashToImportPath(importPath)
				return importPath, gopath, nil
			}
		}
	}

	return "", "", fmt.Errorf("Dir %q not a go package or not in GOPATH", dir)
}

func slashToImportPath(path string) string {
	return strings.Replace(path, `\`, "/", -1)
}

func fileHasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return caseInsensitiveEq(s[:len(prefix)], prefix)
}

func fileTrimPrefix(s, prefix string) string {
	if fileHasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func fileHasSuffix(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	return caseInsensitiveEq(s[len(s)-len(suffix):], suffix)
}

func fileTrimSuffix(s, suffix string) string {
	if fileHasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

var slashSep = filepath.Separator

func fileStringEquals(s1, s2 string) bool {
	if len(s1) == 0 {
		return len(s2) == 0
	}
	if len(s2) == 0 {
		return len(s1) == 0
	}
	r1End := s1[len(s1)-1]
	r2End := s2[len(s2)-1]
	if r1End == '/' || r1End == '\\' {
		s1 = s1[:len(s1)-1]
	}
	if r2End == '/' || r2End == '\\' {
		s2 = s2[:len(s2)-1]
	}
	return caseInsensitiveEq(s1, s2)
}

func caseInsensitiveEq(s1, s2 string) bool {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return strings.EqualFold(s1, s2)
	}
	return s1 == s2
}

// parseGoEnvLine parses a "go env" line into a key value pair.
func parseGoEnvLine(line string) (key, value string, ok bool) {
	// Remove any leading "set " found on windows.
	// Match the name to the env var + "=".
	// Remove any quotes.
	// Return result.
	line = strings.TrimPrefix(line, "set ")
	parts := strings.SplitN(line, "=", 2)
	if len(parts) < 2 {
		return "", "", false
	}

	un, err := strconv.Unquote(parts[1])
	if err != nil {
		return parts[0], parts[1], true
	}
	return parts[0], un, true
}