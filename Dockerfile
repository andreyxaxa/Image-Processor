# Stage 1: Modules caching
FROM golang:1.24-alpine AS modules

WORKDIR /modules

COPY go.mod go.sum ./
RUN go mod download

# Stage 2: Builder
FROM golang:1.24-alpine AS builder

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 \
    go build -o /bin/app ./cmd/app

# Stage 3: Final
FROM scratch

COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "/app" ]