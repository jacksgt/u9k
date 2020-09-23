FROM golang:1.15 AS builder
COPY . /u9k
RUN cd /u9k && \
    echo "Downloading Go modules..." && \
    go mod download && \
    echo "Compiling binary..." && \
    CGO_ENABLED=0 \
    go build -tags 'osusergo netgo' u9k/cmd/server && \
    echo "Stripping binary..." && \
    strip server

FROM scratch
COPY --from=builder /u9k/server /
COPY --from=builder /u9k/static/ /static
COPY --from=builder /u9k/templates/ /templates
COPY --from=builder /u9k/migrations/ /migrations
ENV U9K_LISTEN_ADDR=0.0.0.0 \
    U9K_PORT=3000
ENTRYPOINT ["/server"]
