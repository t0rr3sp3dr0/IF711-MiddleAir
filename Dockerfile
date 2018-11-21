FROM golang:latest

WORKDIR /go/src/github.com/t0rr3sp3dr0/middleair
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["middleair"]
