FROM golang:1.15-buster

WORKDIR /go/src/app
RUN apt-get update && apt-get -y install iptables git bash

# Just copy dependecies file so docker doesnt redownload everything everytime
# the source code changes
COPY go.mod go.sum ./

RUN go mod download

# Copy remaind
COPY . .
RUN go build -v -o app
RUN pwd

CMD ["./app"]