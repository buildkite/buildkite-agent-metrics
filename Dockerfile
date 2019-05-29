FROM segment/chamber:2.3 AS chamber
FROM golang:1.12 as builder
WORKDIR /go/src/github.com/buildkite/buildkite-agent-metrics/
COPY . .
RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o buildkite-agent-metrics .

FROM alpine:3.9
RUN apk update && apk add curl ca-certificates
COPY --from=chamber /chamber /bin/chamber
COPY --from=builder /go/src/github.com/buildkite/buildkite-agent-metrics/buildkite-agent-metrics .
