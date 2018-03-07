FROM golang:1.8

WORKDIR /go/src/shiori
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["shiori", "serve"]
