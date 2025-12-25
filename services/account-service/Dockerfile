# =========================
# Stage 1: Proto Builder
# =========================
FROM golang:1.24-alpine AS proto-builder

RUN apk add --no-cache protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

ENV PATH="/go/bin:${PATH}"

WORKDIR /proto

COPY proto/account/account.proto proto/account/account.proto

RUN protoc \
  --proto_path=proto \
  --go_out=proto \
  --go_opt=paths=source_relative \
  --go-grpc_out=proto \
  --go-grpc_opt=paths=source_relative \
  proto/account/account.proto

# =========================
# Stage 2: Go Builder
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files first (cache optimization)
COPY services/account-service/go.mod services/account-service/go.sum ./
RUN go mod download

# Copy generated proto code from proto-builder
COPY --from=proto-builder /proto/proto ./proto

# Copy service source code
COPY services/account-service/ .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
  -o account-service ./cmd/main.go

# =========================
# Stage 3: Runtime
# =========================
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root

# Copy binary
COPY --from=builder /app/account-service .

# Expose REST + gRPC
EXPOSE 8004 50051

# Run as non-root user
RUN addgroup -g 1000 worker && \
  adduser -D -u 1000 -G worker worker && \
  chown -R worker:worker /root

USER worker

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8004/health || exit 1

CMD ["./account-service"]
