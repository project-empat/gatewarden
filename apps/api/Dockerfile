FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/api ./cmd/api && \
    go build -o /bin/migrate ./cmd/migrate

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/migrate /bin/migrate

# Copy migrations — available at /migrations from the build context
# Also mountable via docker-compose volume
COPY migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/bin/api"]
