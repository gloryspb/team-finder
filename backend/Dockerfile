FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /team-finder-api ./cmd/api

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /team-finder-api /usr/local/bin/team-finder-api
COPY migrations ./migrations
EXPOSE 8080
CMD ["team-finder-api"]
