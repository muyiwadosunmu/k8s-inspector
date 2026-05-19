FROM golang:1.20-alpine AS builder
RUN apk add --no-cache git build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/inspector ./cmd/inspector

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/inspector /usr/local/bin/inspector
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/inspector"]
