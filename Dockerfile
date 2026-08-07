# syntax=docker/dockerfile:1

# Classic multi-stage container image for standard-tools-go.
# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache dependency downloads.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source tree.
COPY . .

# Build a static Linux binary for the server.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o /bin/server ./cmd/server

# Final stage: minimal distroless image.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/server /server

USER nonroot:nonroot

EXPOSE 8080 50051

ENTRYPOINT ["/server"]
