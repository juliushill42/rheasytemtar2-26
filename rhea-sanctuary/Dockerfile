FROM golang:1.21-alpine AS builder
WORKDIR /build
COPY main.go .
RUN go mod init rhea-sanctuary && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o sanctuary main.go

FROM scratch
COPY --from=builder /build/sanctuary /sanctuary
EXPOSE 9103
ENTRYPOINT ["/sanctuary"]
