module github.com/buildkite/buildkite-agent-metrics

require (
	cloud.google.com/go v0.37.2
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895
	github.com/aws/aws-lambda-go v1.6.0
	github.com/aws/aws-sdk-go v1.15.66
	github.com/golang/protobuf v1.2.0
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af // indirect
	github.com/newrelic/go-agent v2.7.0+incompatible
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/client_model v0.0.0-20190115171406-56726106282f
	google.golang.org/genproto v0.0.0-20190401181712-f467c93bbac2
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace github.com/buildkite/buildkite-agent-metrics => github.com/segmentio/buildkite-agent-metrics v0.0.0-20190518062504-6cea804e5f54
