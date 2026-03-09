FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init sentry-relay && go build -o sentry-relay .

FROM alpine:3.19
COPY --from=builder /app/sentry-relay /sentry-relay
EXPOSE 9456
ENTRYPOINT ["/sentry-relay"]
