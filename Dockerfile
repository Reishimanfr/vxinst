# Builder
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Build dependencies
RUN apk add --no-cache gcc musl-dev

COPY . .

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-s -w" -tags=jsoniter -o vxinstagram

# Runner
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Copy the binary and templates from the builder stage
COPY --from=builder /app/vxinstagram .
COPY --from=builder /app/templates ./templates

EXPOSE 8080

ENTRYPOINT ["./vxinstagram"]
