FROM golang:alpine
MAINTAINER MuLin <mulin@bbcclive.com>

WORKDIR /go/src/ddns

ADD . .

RUN go build

CMD ["./ddns"]