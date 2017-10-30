FROM golang:1.9

MAINTAINER ysqi <devysq@gmail.com>

RUN go-wrapper download github.com/ysqi/gcodesharp
RUN go-wrapper install github.com/ysqi/gcodesharp


RUN chmod +x $GOPATH/bin/gcodesharp

WORKDIR $GOPATH