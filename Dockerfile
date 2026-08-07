# syntax=docker/dockerfile:1

# Classic multi-stage container image for standard-tools-go.
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependency downloads.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source tree.
COPY . .

# Build a static Linux binary for the server.
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags='-s -w' -o /usr/local/bin/server ./cmd/server

# Final stage: minimal distroless image.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/local/bin/server /usr/local/bin/server

USER nonroot:nonroot

EXPOSE 8080 50051

ENTRYPOINT ["/usr/local/bin/server"]
