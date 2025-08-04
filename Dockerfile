FROM --platform=$BUILDPLATFORM golang:alpine as builder
ARG TARGETOS
ARG TARGETARCH

# SSL Certificate
RUN apk add -U --no-cache ca-certificates
RUN update-ca-certificates

# Build the binary
COPY . /build
RUN cd /build && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app github.com/Nivl/trakt-netflix/cmd/service && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /app github.com/Nivl/trakt-netflix/cmd/auth

# Create the final image
FROM scratch
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
VOLUME /config
CMD ["/app"]
