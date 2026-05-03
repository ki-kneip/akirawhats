FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# ---- runtime ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/docs ./docs

RUN mkdir -p /app/data

RUN touch .env

EXPOSE 8080

CMD ["./server"]
