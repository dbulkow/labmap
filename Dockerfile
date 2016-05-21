FROM golang:1.6
MAINTAINER David Bulkow <david.bulkow@stratus.com>

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

CMD ["go-wrapper", "run", "-labmap", "/resources/lab.map"]

COPY . /go/src/app
RUN go-wrapper install
