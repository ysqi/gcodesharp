FROM golang

MAINTAINER ysqi <devysq@gmail.com>

RUN go get -u github.com/ysqi/gcodesharp  
RUN go install github.com/ysqi/gcodesharp  
RUN chmod +x $GOPATH/bin/gcodesharp

WORKDIR $GOPATH/src