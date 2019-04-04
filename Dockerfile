FROM alpine:3.8
RUN apk update && apk add curl ca-certificates
COPY bin/buildkite-agent-metrics /usr/local/bin/
EXPOSE 8080 8125
ENTRYPOINT ["/usr/local/bin/buildkite-agent-metrics"]
