FROM golang:1.26.1-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o /app/bin/main ./cmd/api/main.go

FROM alpine:3
WORKDIR /app
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/migrations /app/migrations
ENTRYPOINT ["/app/bin/main"]