FROM golang:1.15-buster

WORKDIR /go/src/app
COPY . .

RUN apt-get update && apt-get -y install iptables git bash
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -v -o app

CMD ["./app"]