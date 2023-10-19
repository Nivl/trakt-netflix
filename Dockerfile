FROM golang:alpine as builder
RUN apk add -U --no-cache ca-certificates
RUN update-ca-certificates

COPY . /build
RUN cd /build && go build -o /app .

FROM scratch
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
VOLUME /config
CMD ["/app"]
