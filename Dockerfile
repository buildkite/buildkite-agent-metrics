# Use this Dockerfile when building the image yourself.
# The Buildkite pipeline is similar, but copies prebuilt binaries instead of
# rebuilding them.
FROM public.ecr.aws/docker/library/golang:1.25.4@sha256:698183780de28062f4ef46f82a79ec0ae69d2d22f7b160cf69f71ea8d98bf25d AS builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM public.ecr.aws/docker/library/alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
EXPOSE 8080 8125
ENTRYPOINT ["./buildkite-agent-metrics"]
