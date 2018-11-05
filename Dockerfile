FROM golang:1.11

WORKDIR /go/src/github.com/buckley-w-david/tdt-anibot
COPY . .

RUN cd bot
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["bot"]
