FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init rhea-orchestra && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o orchestra main.go

FROM scratch
COPY --from=builder /build/orchestra /orchestra
EXPOSE 9100
ENTRYPOINT ["/orchestra"]
