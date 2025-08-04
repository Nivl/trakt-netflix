FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETOS
ARG TARGETARCH

# SSL Certificate
RUN apk add -U --no-cache ca-certificates
RUN update-ca-certificates

# Copy the code needed for building the app
COPY cmd /build/cmd
COPY internal /build/internal

COPY go.mod /build/
COPY go.sum /build/

# Build the binaries
WORKDIR /build
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app github.com/Nivl/trakt-netflix/cmd/service
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app github.com/Nivl/trakt-netflix/cmd/auth

# Create the final image
FROM scratch
USER nonroot
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
VOLUME /config
CMD ["/app"]
