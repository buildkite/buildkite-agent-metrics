# Token

This package is an abstraction layer enabling the retrieval of a Buildkite API token
from different sources.

The current supported sources are:

- AWS Systems Manager (parameter store)
- AWS Secrets Manager
- OS environment variable

## How to run tests ?
All the tests for AWS dependant resources require their corresponding auto-generated mocks. Thus,
before running them you need to generate such mocks by executing:

```bash
go generate token/secretsmanager_test.go
go generate token/ssm_test.go
```

