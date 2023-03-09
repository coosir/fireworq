FROM golang:1.17.8 as builder
ENV APP_DIR /go/src/github.com/coosir/middleman

WORKDIR ${APP_DIR}
COPY . .
RUN make release PRERELEASE=

FROM alpine:3.15.4
ENV APP_DIR /go/src/github.com/coosir/middleman

COPY --from=builder ${APP_DIR}/middleman /usr/local/bin/
ENV MIDDLEMAN_BIND 0.0.0.0:8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/middleman"]
