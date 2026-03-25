# Stage 1: Build the binary
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/bin/rtmx ./cmd/rtmx

# Stage 2: Minimal runtime image
FROM alpine:latest

RUN apk add --no-cache ca-certificates git
COPY --from=builder /usr/bin/rtmx /usr/bin/rtmx

ENTRYPOINT ["rtmx"]
