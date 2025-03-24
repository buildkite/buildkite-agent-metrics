# Use this Dockerfile when building the image yourself.
# The Buildkite pipeline is similar, but copies prebuilt binaries instead of
# rebuilding them.
FROM public.ecr.aws/docker/library/golang:1.24.1@sha256:52ff1b35ff8de185bf9fd26c70077190cd0bed1e9f16a2d498ce907e5c421268 AS builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM public.ecr.aws/docker/library/alpine:3.21.0@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
EXPOSE 8080 8125
ENTRYPOINT ["./buildkite-agent-metrics"]
