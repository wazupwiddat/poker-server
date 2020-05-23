FROM golang:alpine AS builder

RUN mkdir /build
ADD . /build
WORKDIR /build

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/poker-server cmd/main.go

FROM scratch
COPY --from=builder /build/bin /app
COPY --from=builder /build/server/templates /app/templates
COPY --from=builder /build/server/assets /app/assets
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app

# Command to run
CMD ["./poker-server"]