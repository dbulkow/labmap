FROM golang:1.6
MAINTAINER David Bulkow <david.bulkow@stratus.com>

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

CMD ["go-wrapper", "run", "-map", "/resources/lab.map"]

COPY . /go/src/app
COPY lab.map /resources/lab.map
RUN go-wrapper install
