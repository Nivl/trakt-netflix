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
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /service github.com/Nivl/trakt-netflix/cmd/service
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /auth github.com/Nivl/trakt-netflix/cmd/auth

RUN adduser -u 10000 -SH -s /bin/false nonroot

# Create the final image
FROM scratch

COPY --from=builder /service /service
COPY --from=builder /auth /auth
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
VOLUME /config

COPY --from=builder /etc/passwd /etc/passwd
USER nonroot

CMD ["/service"]
