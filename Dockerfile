FROM golang:alpine as builder

ENV PACKAGE github.com/travis-g/draas

RUN apk add --no-cache git

COPY . $GOPATH/src/$PACKAGE/
WORKDIR $GOPATH/src/$PACKAGE/

ENV GOOS=linux GOARCH=amd64 CGO_ENABLED=0
RUN go get -d -v && go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/draas

FROM scratch

COPY --from=builder /go/bin/draas /usr/bin/draas

EXPOSE 8000

ENTRYPOINT ["/usr/bin/draas"]
