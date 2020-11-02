FROM golang:1.15-alpine

WORKDIR /go/src/app
COPY . .

RUN apk update && apk add iptables git bash
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]