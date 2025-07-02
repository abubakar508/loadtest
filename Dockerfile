FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o loadtester ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/loadtester .
EXPOSE 8080
CMD ["./loadtester"]
