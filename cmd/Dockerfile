FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/schroebie-backend/
COPY . .
RUN go get -d -v all
RUN cd cmd/ && GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/main


FROM alpine:latest

WORKDIR /app
COPY --from=builder /go/bin/main /go/bin/main
EXPOSE 8080
ENTRYPOINT ["/go/bin/main"]