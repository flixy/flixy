FROM golang

ADD . /go/src/github.com/flixy/flixy
RUN go install github.com/flixy/flixy
ENTRYPOINT /go/bin/flixy

EXPOSE 80
