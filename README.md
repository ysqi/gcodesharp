# GCodeSharp

GCodeSharp is a CLI library for Go language code review applications.
This application is a tool to generate the report to quickly review the golang code.

# Install
```shell
go get -u github.com/ysqi/gcodesharp
```

# TODO

- [x] support go test
- [x] support go test result report to junit
- [ ] support go vet
- [ ] support go fmt
- [ ] support golint
- [ ] create a html report contain all thing

# Get Help
you need run application with args `-h`(-help) to get help.
you can add issue to ask me.
```shell
gcodesharp -h
```

# Get Junit Report

gcodesharp support more one golang project package path . default is current dir if not set.

```shell
gcodesharp --junit=$HOME/tmp/myjunit.xml github.com/ysqi/gcodesharp github.com/ysqi/com
```
this command will run go test for current dir and all child dir. Is actually equivalent to the following command:
```shell
go test -cover -timeout 30s -v github.com/ysqi/gcodesharp... github.com/ysqi/com...
```
`github.com/ysqi/gcodesharp...` mean contains import path prefixed with `github.com/ysqi/gcodesharp`.