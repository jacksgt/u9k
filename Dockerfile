FROM golang:1.15 AS builder
COPY . /u9k
RUN cd /u9k && \
    go mod download && \
    go mod verify && \
    CGO_ENABLED=0 \
    go build -tags 'osusergo netgo' u9k/cmd/server && \
    strip server

FROM scratch
COPY --from=builder /u9k/server /
COPY --from=builder /u9k/static /static
ENV LISTEN_ADDR=0.0.0.0 \
    PORT=3000
ENTRYPOINT ["/server"]
