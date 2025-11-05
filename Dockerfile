FROM golang:1.25.3 AS builder
WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/s3manager ./cmd/s3manager && \
    chmod +x /build/s3manager

FROM alpine:3.22.1
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
WORKDIR /app
COPY --from=builder /build/s3manager /app/s3manager
CMD ["./s3manager"]
