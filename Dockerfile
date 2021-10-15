FROM golang:1.17.2 as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY ./main.go  ./

RUN go build \
    -o /go/bin/main \
    -ldflags '-s -w'

FROM alpine:3.14.2 as runner

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/main /app/main

ENTRYPOINT ["/app/main"]
