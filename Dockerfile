FROM golang:alpine AS builder
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o broker ./cmd/api

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/broker .
CMD ["./broker"]
