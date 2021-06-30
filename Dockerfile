FROM gcr.io/distroless/base

LABEL maintainer="Rajat Vig <rajat.vig@gmail.com>"

COPY dist/buildkite-agent-metrics-linux-amd64 /bin/buildkite-agent-metrics

EXPOSE      8080
ENTRYPOINT  [ "/bin/buildkite-agent-metrics" ]
CMD         [ "-backend", "-prometheus" ]
