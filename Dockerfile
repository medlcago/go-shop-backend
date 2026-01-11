FROM golang:1.25-alpine as builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /app/bin/main ./cmd/api/main.go

FROM alpine:3
WORKDIR /app
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/migrations /app/migrations
ENTRYPOINT ["/app/bin/main"]