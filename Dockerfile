FROM golang:latest

ADD . /go/src/github.com/Pigmice2733/scouting-backend/
WORKDIR /go/src/github.com/Pigmice2733/scouting-backend/

RUN go get -t -v ./...

ENTRYPOINT [ "go", "test", "./..." ]
