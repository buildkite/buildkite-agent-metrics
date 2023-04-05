FROM buildkite/agent:3.45.0 AS agent

FROM public.ecr.aws/docker/library/golang:1.20.2-alpine3.17
COPY --from=agent /usr/local/bin/buildkite-agent /usr/local/bin/buildkite-agent
RUN apk --no-cache add bash zip curl aws-cli github-cli
