FROM public.ecr.aws/docker/library/golang:1.23.3@sha256:73f06be4578c9987ce560087e2e2ea6485fb605e3910542cadd8fa09fc5f3e31 as builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM public.ecr.aws/docker/library/alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d
RUN apk update && apk add curl ca-certificates
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
EXPOSE 8080 8125
ENTRYPOINT ["./buildkite-agent-metrics"]
