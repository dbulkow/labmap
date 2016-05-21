FROM golang:1.6
MAINTAINER David Bulkow <david.bulkow@stratus.com>

CMD ["go-wrapper", "run", "-labmap", "/resources/lab.map"]

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ONBUILD COPY . /go/src/app
ONBUILD RUN go-wrapper install
