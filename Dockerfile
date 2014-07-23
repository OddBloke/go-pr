FROM ubuntu:latest
MAINTAINER Daniel Watkins <daniel@daniel-watkins.co.uk>

RUN apt-get update
RUN apt-get install -y golang git gcc

ENV GOPATH /go
RUN go get github.com/OddBloke/go-pr
WORKDIR /go/src/github.com/OddBloke/go-pr
ADD . /go/src/github.com/OddBloke/go-pr
RUN go get
RUN go build

EXPOSE 8123

ENTRYPOINT ./go-pr
