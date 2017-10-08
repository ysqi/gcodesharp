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
- [x] support go fmt
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
gofmt -d -e  [all go files of $GOPATH/src/github.com/ysqi/gcodesharp]
```
`github.com/ysqi/gcodesharp...` mean contains import path prefixed with `github.com/ysqi/gcodesharp`.

**Note**: gofmt result as a part of go test. and the result will write to junit file as a testsuite, such like this:
```xml
<testsuite tests="226" failures="0" errors="0" time="2.4243922" name="/usr/local/go/bin/gofmt" timestamp="2017-10-08T23:39:45">
	<properties>
		<property name="go.version" value="go1.8"></property>
		<property name="os" value="darwin"></property>
		<property name="arch" value="amd64"></property>
	</properties>
	<testcase classname="gofmt" name="gfmt/gfmt.go" time="0"></testcase>
	...
<testsuite>
```