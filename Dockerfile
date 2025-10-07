# Use this Dockerfile when building the image yourself.
# The Buildkite pipeline is similar, but copies prebuilt binaries instead of
# rebuilding them.
FROM public.ecr.aws/docker/library/golang:1.25.1@sha256:d7098379b7da665ab25b99795465ec320b1ca9d4addb9f77409c4827dc904211 AS builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM public.ecr.aws/docker/library/alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
EXPOSE 8080 8125
ENTRYPOINT ["./buildkite-agent-metrics"]
