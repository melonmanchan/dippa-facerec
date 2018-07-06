FROM golang:1.10

ADD . /go/src
WORKDIR /go/src
RUN go get ./...
RUN go build -o main ./src/main.go
CMD ["/go/src/main"]
