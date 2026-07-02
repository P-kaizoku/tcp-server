FROM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY lb/main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o lb main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lb .
EXPOSE 9000
CMD ["./lb"]
