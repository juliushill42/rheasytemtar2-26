FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init rhea-cloning && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o cloning main.go

FROM scratch
COPY --from=builder /build/cloning /cloning
EXPOSE 9102
ENTRYPOINT ["/cloning"]
