ARG BUILD_TIME=""
ARG GIT_COMMIT=""

FROM golang:1.20-alpine AS builder
ARG BUILD_TIME
ARG GIT_COMMIT
RUN apk add --no-cache git build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Embed build metadata into the binary using -ldflags
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags "-s -w -X 'main.buildTime=${BUILD_TIME}' -X 'main.gitCommit=${GIT_COMMIT}'" \
    -o /app/inspector ./cmd/inspector

FROM alpine:3.18
ARG BUILD_TIME
ARG GIT_COMMIT
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/inspector /usr/local/bin/inspector
LABEL org.opencontainers.image.created=$BUILD_TIME
LABEL org.opencontainers.image.revision=$GIT_COMMIT
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/inspector"]
