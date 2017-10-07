package gtest

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestParse(t *testing.T) {

	tpl, err := template.New("packagecontent").Parse(contentTpl)
	if err != nil {
		t.Fatal(err)
	}
	var content bytes.Buffer

	testcases := []string{
		"pass.txt",
		"fail.txt",
		"skip.txt",
		"go_1_4.txt",
		"go_1_5.txt",
		"go_1_7.txt",
		"mixed.txt",
		"parallel.txt",
		"coverage.txt",
		"multipkg-coverage.txt",
		"syntax-error.txt",
		"panic.txt",
		"empty.txt",
		"race.txt",
	}
	for _, c := range testcases {
		file, err := os.Open(filepath.Join("./testdata", c))
		if err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		pkgs, err := parse(scanner, false)
		if err != nil {
			t.Fatal(err)
		}
		content.Reset()
		if err := tpl.Execute(&content, pkgs); err != nil {
			t.Fatal(err)
		}
		wantFile := filepath.Join("./testdata", c[:strings.LastIndex(c, ".")]+".result.txt")
		data, err := ioutil.ReadFile(wantFile)
		if err != nil {
			t.Fatal(err)
		}
		want := strings.Trim(string(data), "\n")
		got := strings.Trim(content.String(), "\n")
		if want != got {
			t.Fatalf("%s\nwant:\n-----start------\n%s\n-----end------\nbut got:\n-----start------\n%s\n-----end------\n", c, want, got)
		}
	}
}

var contentTpl = `{{range .}}package {{.Name}} test {{if .Failed}}failed{{else}}passed{{end}}
Coverage: {{if eq .Coverage -1.0}}unset{{else}}{{printf "%.2f" .Coverage}}%{{end}}
Cost: {{printf "%.3f" .Cost}} second
Pass: {{.PassCount}}, Fail: {{.FailCount}}, Skip: {{.SkipCount}}
Failed cause:{{.Err}}
Tests:
{{range .Units}}	+{{.Result}}	{{.Name}}	Spend time={{printf "%.3f" .Cost}} sencond	Output:{{if .Output}}
{{.Output}}{{else}}<nil>{{end}}
{{end}}
{{end}}
`
