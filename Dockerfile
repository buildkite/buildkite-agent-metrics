FROM public.ecr.aws/docker/library/golang:1.23.0@sha256:613a108a4a4b1dfb6923305db791a19d088f77632317cfc3446825c54fb862cd as builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM public.ecr.aws/docker/library/alpine:3.20.1@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0
RUN apk update && apk add curl ca-certificates
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
EXPOSE 8080 8125
ENTRYPOINT ["./buildkite-agent-metrics"]
