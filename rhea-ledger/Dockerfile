FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init rhea-ledger && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o ledger main.go

FROM scratch
COPY --from=builder /build/ledger /ledger
EXPOSE 9104
ENTRYPOINT ["/ledger"]
