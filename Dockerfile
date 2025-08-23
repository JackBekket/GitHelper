FROM golang:latest AS builder
LABEL maintainer="mintyleaf <mintyleafdev@gmail.com>"

WORKDIR /build

COPY go.mod go.sum main.go ./
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o githellper ./main.go

FROM alpine
WORKDIR /

COPY --from=builder /build/githellper ./githellper

CMD ["/githellper"]
