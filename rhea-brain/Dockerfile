FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init rhea-brain && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o brain main.go

FROM scratch
COPY --from=builder /build/brain /brain
EXPOSE 9101
ENTRYPOINT ["/brain"]
