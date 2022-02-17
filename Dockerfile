FROM docker.io/library/golang:1.17 AS builder

WORKDIR /u9k

# Enable Go's DNS resolver to read from /etc/hosts
# see https://github.com/golang/go/issues/22846
RUN echo "hosts: files dns" > /etc/nsswitch.conf.min
# Create a minimal passwd so we can run as non-root in the container
RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc/passwd.min
RUN mkdir -p /faketmp

# Only download Go modules (improves build caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy our source code over and build the binary
COPY . .
RUN CGO_ENABLED=0 \
    go build -ldflags='-s -w' -tags 'osusergo netgo' u9k/cmd/server

FROM scratch

ENV U9K_LISTEN_ADDR=0.0.0.0 \
    U9K_PORT=3000

COPY --from=builder /u9k/server /
COPY --from=builder /u9k/static/ /static
COPY --from=builder /u9k/templates/ /templates
COPY --from=builder /u9k/migrations/ /migrations
COPY --from=builder /etc/nsswitch.conf.min /etc/nsswitch.conf
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd.min /etc/passwd
# Make sure we have a temp directory in the container (not allowed to create it as non-root)
COPY --from=builder --chown=nobody:0 /faketmp /tmp

USER nobody

ENTRYPOINT ["/server"]
