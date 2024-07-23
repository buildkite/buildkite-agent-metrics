# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [v5.9.8](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.8) (2024-07-22)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.7...v5.9.8)

### Fixed
- fix to remove reset in prometheus backend [#296](https://github.com/buildkite/buildkite-agent-metrics/pull/296) (@wolfeidau)

### Changed
- Specify storage resolution [#125](https://github.com/buildkite/buildkite-agent-metrics/pull/125) (@patrobinson)

## [v5.9.7](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.7) (2024-07-04)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.6...v5.9.7)

### Fixed
- fix for prometheus backend panic'ing when using unclustered agents [#288](https://github.com/buildkite/buildkite-agent-metrics/pull/288) (@wolfeidau)

### Dependencies
- Update the lambda dockerfile to use al2023 [#291](https://github.com/buildkite/buildkite-agent-metrics/pull/291) (@wolfeidau)

### Internal
- Allow dependabot to keep our dockerfiles up to date [#289](https://github.com/buildkite/buildkite-agent-metrics/pull/289) (@wolfeidau)

## [v5.9.6](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.6) (2024-06-05)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.5...v5.9.6)

### Changed
- This change enables configuration of http connection pooling [#286](https://github.com/buildkite/buildkite-agent-metrics/pull/286) (@wolfeidau)
- This change will ensure a single HTTP client is created [#285](https://github.com/buildkite/buildkite-agent-metrics/pull/285) (@wolfeidau)

## [v5.9.5](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.5) (2024-05-27)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.4...v5.9.5)

### Changed
- Adding more debug traces for HTTP and enhance security [#283](https://github.com/buildkite/buildkite-agent-metrics/pull/283) (@CheeseStick)
- Fix CloudWatch region defaulting [#282](https://github.com/buildkite/buildkite-agent-metrics/pull/282) (@DrJosh9000)

### Dependencies
- build(deps): bump cloud.google.com/go/monitoring from 1.17.0 to 1.19.0 [#280](https://github.com/buildkite/buildkite-agent-metrics/pull/280) (@dependabot[bot])
- build(deps): bump github.com/prometheus/client_model from 0.6.0 to 0.6.1 [#276](https://github.com/buildkite/buildkite-agent-metrics/pull/276) (@dependabot[bot])

## [v5.9.4](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.4) (2024-05-06)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.3...v5.9.4)

### Changed
- Log DNS resolution times when debug http is enabled [#278](https://github.com/buildkite/buildkite-agent-metrics/pull/278) (@patrobinson)

### Dependencies
- build(deps): bump golang.org/x/net from 0.20.0 to 0.23.0 [#275](https://github.com/buildkite/buildkite-agent-metrics/pull/275) (@dependabot[bot])
- build(deps): bump github.com/prometheus/client_golang from 1.17.0 to 1.19.0 [#269](https://github.com/buildkite/buildkite-agent-metrics/pull/269) (@dependabot[bot])
- build(deps): bump github.com/aws/aws-sdk-go from 1.51.1 to 1.51.21 [#274](https://github.com/buildkite/buildkite-agent-metrics/pull/274) (@dependabot[bot])
- build(deps): bump github.com/prometheus/client_model from 0.5.0 to 0.6.0 [#267](https://github.com/buildkite/buildkite-agent-metrics/pull/267) (@dependabot[bot])
- build(deps): bump github.com/aws/aws-sdk-go from 1.50.35 to 1.51.1 [#268](https://github.com/buildkite/buildkite-agent-metrics/pull/268) (@dependabot[bot])
- build(deps): bump github.com/aws/aws-lambda-go from 1.42.0 to 1.46.0 [#259](https://github.com/buildkite/buildkite-agent-metrics/pull/259) (@dependabot[bot])
- build(deps): bump github.com/aws/aws-sdk-go from 1.48.16 to 1.50.35 [#265](https://github.com/buildkite/buildkite-agent-metrics/pull/265) (@dependabot[bot])
- build(deps): bump google.golang.org/protobuf from 1.31.0 to 1.33.0 [#266](https://github.com/buildkite/buildkite-agent-metrics/pull/266) (@dependabot[bot])

## [v5.9.3](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.3) (2023-12-19)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.2...v5.9.3)

### Changed
- Add v5 to module path [#248](https://github.com/buildkite/buildkite-agent-metrics/pull/248) (@DrJosh9000)

### Dependencies
- build(deps): bump golang.org/x/crypto from 0.14.0 to 0.17.0 [#246](https://github.com/buildkite/buildkite-agent-metrics/pull/246) (@dependabot[bot])
- build(deps): bump cloud.google.com/go/monitoring from 1.16.3 to 1.17.0 [#245](https://github.com/buildkite/buildkite-agent-metrics/pull/245) (@dependabot[bot])
- build(deps): bump github.com/aws/aws-lambda-go from 1.41.0 to 1.42.0 [#244](https://github.com/buildkite/buildkite-agent-metrics/pull/244) (@dependabot[bot])

## [v5.9.2](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.2) (2023-12-12)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.1...v5.9.2)

### Fixed
- Fix non-Secrets Manager token providers [#243](https://github.com/buildkite/buildkite-agent-metrics/pull/243) (@DrJosh9000)

### Changed
- Allow env vars to control debug logging for the lambda [#238](https://github.com/buildkite/buildkite-agent-metrics/pull/238) (@triarius)

### Dependencies
- Bump github.com/aws/aws-sdk-go from 1.48.3 to 1.48.4 to 1.48.16 [#237](https://github.com/buildkite/buildkite-agent-metrics/pull/237), [#241](https://github.com/buildkite/buildkite-agent-metrics/pull/241) (@dependabot[bot])

## [v5.9.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.1) (2023-11-27)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.9.0...v5.9.1)

### Changed
- Support for multiple secrets manager secrets command seperated [#233](https://github.com/buildkite/buildkite-agent-metrics/pull/233) (@lucylura)

### Fixed
- Ignore Cluster label/dimension/tag for empty unclustered queues. This may fix continuity errors when clusters are not used [#234](https://github.com/buildkite/buildkite-agent-metrics/pull/234) (@triarius)

### Internal
- Document SSM Parameters names may be comma separated [#235](https://github.com/buildkite/buildkite-agent-metrics/pull/235) (@triarius)

### Dependencies
- build(deps): bump github.com/aws/aws-sdk-go from 1.47.3 to 1.48.3 [#232](https://github.com/buildkite/buildkite-agent-metrics/pull/232) (@dependabot[bot])

## [v5.9.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.9.0) (2023-11-22)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.8.0...v5.9.0)

> [!WARNING]
> This release adds a new Cluster label/tag/dimension, which is populated when using agent cluster tokens. This may break continuity with existing time series!

### Added
- Collect from multiple clusters [#227](https://github.com/buildkite/buildkite-agent-metrics/pull/227) (@DrJosh9000)
- feat(gcp): add env vars for buildkite queues and gcp project id [#212](https://github.com/buildkite/buildkite-agent-metrics/pull/212) (@NotArpit)

### Fixed
- Change build process to better support `provided.al2` [#225](https://github.com/buildkite/buildkite-agent-metrics/pull/225) (@triarius)
- fix(collector): exit on 401 when queues specified [#211](https://github.com/buildkite/buildkite-agent-metrics/pull/211) (@NotArpit)
- Fix another reference to go1.x [#230](https://github.com/buildkite/buildkite-agent-metrics/pull/230) (@jradtilbrook)

### Internal
- Split Collect [#226](https://github.com/buildkite/buildkite-agent-metrics/pull/226) (@DrJosh9000)
- Various dependency updates [#206](https://github.com/buildkite/buildkite-agent-metrics/pull/206), [#208](https://github.com/buildkite/buildkite-agent-metrics/pull/208), [#213](https://github.com/buildkite/buildkite-agent-metrics/pull/213), [#215](https://github.com/buildkite/buildkite-agent-metrics/pull/215), [#216](https://github.com/buildkite/buildkite-agent-metrics/pull/216), [#217](https://github.com/buildkite/buildkite-agent-metrics/pull/217), [#218](https://github.com/buildkite/buildkite-agent-metrics/pull/218), [#219](https://github.com/buildkite/buildkite-agent-metrics/pull/219), [#220](https://github.com/buildkite/buildkite-agent-metrics/pull/220), [#221](https://github.com/buildkite/buildkite-agent-metrics/pull/221), [#222](https://github.com/buildkite/buildkite-agent-metrics/pull/222), [#223](https://github.com/buildkite/buildkite-agent-metrics/pull/223) (@dependabot[bot])

## [v5.8.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.8.0) (2023-09-15)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.7.0...v5.8.0)

### Changed
- Exit with code 4 on 401 response from the API [#203](https://github.com/buildkite/buildkite-agent-metrics/pull/203) (@NotArpit)
- Bump github.com/aws/aws-sdk-go to 1.45.6 [#191](https://github.com/buildkite/buildkite-agent-metrics/pull/191) [#194](https://github.com/buildkite/buildkite-agent-metrics/pull/194) [#195](https://github.com/buildkite/buildkite-agent-metrics/pull/195) [#197](https://github.com/buildkite/buildkite-agent-metrics/pull/197) [#198](https://github.com/buildkite/buildkite-agent-metrics/pull/198) [#199](https://github.com/buildkite/buildkite-agent-metrics/pull/199) [#201](https://github.com/buildkite/buildkite-agent-metrics/pull/201) (@dependabot[bot])
- Bump github.com/newrelic/go-agent to 3.24.1+incompatible [#193](https://github.com/buildkite/buildkite-agent-metrics/pull/193) [#196](https://github.com/buildkite/buildkite-agent-metrics/pull/196) (@dependabot[bot])

## [v5.7.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.7.0) (2023-07-24)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.6.0...v5.7.0)

### Changed
- Make the timeout configurable [#184](https://github.com/buildkite/buildkite-agent-metrics/pull/184) (@mcncl)
- Update the role ARN used during releases [#162](https://github.com/buildkite/buildkite-agent-metrics/pull/162) (@yob)
- Many dependency version bumps (@dependabot[bot])

## [v5.6.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.6.0) (2023-04-11)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.6...v5.6.0)

### Changed
- Bump github.com/aws/aws-sdk-go from 1.44.234 to 1.44.239 [#157](https://github.com/buildkite/buildkite-agent-metrics/pull/157) (@dependabot[bot])

### Fixed
- Handle API errors when querying queue [#139](https://github.com/buildkite/buildkite-agent-metrics/pull/139) (@dyson)

## [v5.5.6](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.6) (2023-04-10)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.5...v5.5.6)

### Changed
- Remove comments in the middle of a bash command in the release script [#155](https://github.com/buildkite/buildkite-agent-metrics/pull/155) (@triarius)

## [v5.5.5](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.5) (2023-04-10)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.4...v5.5.5)

### Changed
- add notes about what the job states mean [#130](https://github.com/buildkite/buildkite-agent-metrics/pull/130) (@edmund-huber)
- More fixes to the automated release [#153](https://github.com/buildkite/buildkite-agent-metrics/pull/153) (@triarius)

## [v5.5.4](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.4) (2023-04-10)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.3...v5.5.4)

### Changed
- Fix `--verify-tag` not available in github-cli 2.20 for release automation [#151](https://github.com/buildkite/buildkite-agent-metrics/pull/151) (@triarius)

## [v5.5.3](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.3) (2023-04-10)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.2...v5.5.3)

### Changed
- More fixes to release automation [#149](https://github.com/buildkite/buildkite-agent-metrics/pull/149) (@triarius)

## [v5.5.2](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.2) (2023-04-09)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.1...v5.5.2)

### Changed
- Attempt to fix release process [#147](https://github.com/buildkite/buildkite-agent-metrics/pull/147) (@triarius)

## [v5.5.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.1) (2023-04-05)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.5.0...v5.5.1)

### Changed
- Update release process to generate checksums [#145](https://github.com/buildkite/buildkite-agent-metrics/pull/145) (@triarius)
- Allow dependabot to slowly keep gomod up to date [#135](https://github.com/buildkite/buildkite-agent-metrics/pull/135) (@yob)
- Dependency updates (@dependabot[bot])

## [v5.5.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.5.0) (2023-03-16)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.4.0...v5.5.0)

### Changed
- Fixed release process with OIDC [#133](https://github.com/buildkite/buildkite-agent-metrics/pull/133) [#134](https://github.com/buildkite/buildkite-agent-metrics/pull/134) (@yob)
- Update Go (1.20), Alpine (3.17), and all modules [#131](https://github.com/buildkite/buildkite-agent-metrics/pull/131) (@DrJosh9000)

## [v5.4.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.4.0) (2022-06-10)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.3.0...v5.4.0)

### Changed
- Standardize http.Client collector configurations [#121](https://github.com/buildkite/buildkite-agent-metrics/pull/121) (@alloveras)
- Update AWS Lambda SDK v1.6.0 -> v1.28.0, add a lambda-specific dockerfile [#120](https://github.com/buildkite/buildkite-agent-metrics/pull/120) (@ohookins)

## [v5.3.0](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.2.1...v5.3.0) (2021-07-16)

### Addded

* Support reading an agent token from the environment [#116](https://github.com/buildkite/buildkite-agent-metrics/pull/116) ([@cole-h](https://github.com/cole-h))

## [v5.2.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.2.0) (2021-07-01)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.2.0...v5.2.1)

### Added

* Support for more AWS Regions (af-south-1, ap-east-1, ap-southeast-2, ap-southeast-1, eu-south-1, me-south-1) [#109](https://github.com/buildkite/buildkite-agent-metrics/pull/109)
* ARM64 binaries for Linux and macOS

### Changed

* Build using golang 1.16
* Update newrelic/go-agent from v2.7.0 to v3.0.0 [#111](https://github.com/buildkite/buildkite-agent-metrics/pull/111) (@mallyvai)

## [v5.2.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.2.0) (2020-03-05)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.1.0...v5.2.0)

### Changed
- Add support for AWS SecretsManager as BK token provider [#98](https://github.com/buildkite/buildkite-agent-metrics/pull/98) (@alloveras)
- Don't exit on when error is encountered [#94](https://github.com/buildkite/buildkite-agent-metrics/pull/94) (@amalucelli)
- Stackdriver: Use organization specific metric names. [#87](https://github.com/buildkite/buildkite-agent-metrics/pull/87) (@philwo)
- Fix typo in README.md. [#88](https://github.com/buildkite/buildkite-agent-metrics/pull/88) (@philwo)

## [v5.1.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.1.0) (2019-05-18)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v5.0.0...v5.1.0)

### Changed
- Support multiple queue params [#86](https://github.com/buildkite/buildkite-agent-metrics/pull/86) (@lox)
- Add New Relic backend [#85](https://github.com/buildkite/buildkite-agent-metrics/pull/85) (@chloehutchinson)
- Add Stackdriver Backend [#78](https://github.com/buildkite/buildkite-agent-metrics/pull/78) (@winfieldj)

## [v5.0.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v5.0.0) (2019-05-05)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v4.1.2...v5.0.0)

### Changed
- Add BusyAgentPercentage metric [#80](https://github.com/buildkite/buildkite-agent-metrics/pull/80) (@arromer)
- Drop metrics with only queue dimension [#82](https://github.com/buildkite/buildkite-agent-metrics/pull/82) (@lox)
- Add WaitingJobsCount metric [#81](https://github.com/buildkite/buildkite-agent-metrics/pull/81) (@lox)
- Read AWS_REGION for cloudwatch, default to us-east-1 [#79](https://github.com/buildkite/buildkite-agent-metrics/pull/79) (@lox)
- Add a Dockerfile [#77](https://github.com/buildkite/buildkite-agent-metrics/pull/77) (@amalucelli)
- Enforce Buildkite-Agent-Metrics-Poll-Duration header [#83](https://github.com/buildkite/buildkite-agent-metrics/pull/83) (@lox)
- Add support for reading buildkite token from ssm [#76](https://github.com/buildkite/buildkite-agent-metrics/pull/76) (@arromer)
- Update bucket publishing for new regions [#74](https://github.com/buildkite/buildkite-agent-metrics/pull/74) (@lox)
- Update the readme to have the correct Environment variables and expla… [#73](https://github.com/buildkite/buildkite-agent-metrics/pull/73) (@bmbentson)

## [v4.1.3](https://github.com/buildkite/buildkite-agent-metrics/tree/v4.1.3) (2019-03-26)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v4.1.2...v4.1.3)

### Changed
- Update bucket publishing for new regions [#74](https://github.com/buildkite/buildkite-agent-metrics/pull/74) (@lox)
- Update the readme to have the correct Environment variables and expla… [#73](https://github.com/buildkite/buildkite-agent-metrics/pull/73) (@bmbentson)

## [v4.1.2](https://github.com/buildkite/buildkite-agent-metrics/tree/v4.1.2) (2019-01-21)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v4.1.1...v4.1.2)

### Fixed
- Add back cloudwatch metric with only Queue dimension [#69](https://github.com/buildkite/buildkite-agent-metrics/pull/69) (@lox)

## [v4.1.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v4.1.1) (2019-01-21)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v4.1.0...v4.1.1)

### Fixed
- Add missing organization dimension to per-queue metrics [#68](https://github.com/buildkite/buildkite-agent-metrics/pull/68) (@lox)

## [v4.1.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v4.1.0) (2019-01-03)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v4.0.0...v4.1.0)

### Changed
- Expose org slug as a cloudwatch dimension [#67](https://github.com/buildkite/buildkite-agent-metrics/pull/67) (@lox)
- Clarify lambda handler in README, add example [#66](https://github.com/buildkite/buildkite-agent-metrics/pull/66) (@lox)

## [v4.0.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v4.0.0) (2018-11-01)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v3.1.0...v4.0.0)

### Changed
- Update dependencies [#62](https://github.com/buildkite/buildkite-agent-metrics/pull/62) (@lox)
- Move to golang 1.11 [#61](https://github.com/buildkite/buildkite-agent-metrics/pull/61) (@lox)
- Move to aws lambda go [#60](https://github.com/buildkite/buildkite-agent-metrics/pull/60) (@lox)
- Remove unused vendors [#57](https://github.com/buildkite/buildkite-agent-metrics/pull/57) (@paulolai)
- Update references to  github.com/buildkite/buildkite-metrics [#56](https://github.com/buildkite/buildkite-agent-metrics/pull/56) (@paulolai)
- Update readme to reflect elastic stack's changed paths [#54](https://github.com/buildkite/buildkite-agent-metrics/pull/54) (@lox)
- Update capitalization  on Datadog [#52](https://github.com/buildkite/buildkite-agent-metrics/pull/52) (@irabinovitch)

## [v3.1.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v3.1.0) (2018-08-17)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v3.0.1...v3.1.0)

### Changed
- Add a 5 second timeout for metrics requests [#50](https://github.com/buildkite/buildkite-agent-metrics/pull/50) (@lox)
- Improve running docs [#49](https://github.com/buildkite/buildkite-agent-metrics/pull/49) (@lox)
- Allow a custom cloudwatch dimension flag [#46](https://github.com/buildkite/buildkite-agent-metrics/pull/46) (@lox)

## [v3.0.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v3.0.1) (2018-07-12)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v3.0.0...v3.0.1)

### Changed
- Reset prometheus queue gauges to prevent stale values persisting [#45](https://github.com/buildkite/buildkite-agent-metrics/pull/45) (@majolo)

## [v3.0.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v3.0.0) (2018-04-17)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v2.1.0...v3.0.0)

### Changed
- Update buildkite-metrics to use the agent metrics api [#40](https://github.com/buildkite/buildkite-agent-metrics/pull/40) (@sj26)

## [v2.1.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v2.1.0) (2018-03-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v2.0.2...v2.1.0)

### Changed
- Add Prometheus metrics backend [#39](https://github.com/buildkite/buildkite-agent-metrics/pull/39) (@martinbaillie)
- Ensure statsd commands are flushed after each run [#38](https://github.com/buildkite/buildkite-agent-metrics/pull/38) (@theist)
- Small typo in readme [#35](https://github.com/buildkite/buildkite-agent-metrics/pull/35) (@theist)

## [v2.0.2](https://github.com/buildkite/buildkite-agent-metrics/tree/v2.0.2) (2018-01-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v2.0.1...v2.0.2)

Skipped version due to release issues.

## [v2.0.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v2.0.1) (2018-01-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v2.0.0...v2.0.1)

Skipped version due to release issues.

## [v2.0.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v2.0.0) (2017-11-27)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.6.0...v2.0.0)

### Changed

## [v1.6.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.6.0) (2017-11-22)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.5.0...v1.6.0)

### Changed
- Add an endpoint and better user-agent information [#34](https://github.com/buildkite/buildkite-agent-metrics/pull/34) (@lox)
- Punycode pipeline names [#33](https://github.com/buildkite/buildkite-agent-metrics/pull/33) (@rbvigilante)

## [v1.5.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.5.0) (2017-08-11)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.4.2...v1.5.0)

### Changed
- Add retry for failed bk calls to lambda [#30](https://github.com/buildkite/buildkite-agent-metrics/pull/30) (@lox)

## [v1.4.2](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.4.2) (2017-03-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.4.1...v1.4.2)

### Changed
- Add BUILDKITE_QUIET support to lambda [#28](https://github.com/buildkite/buildkite-agent-metrics/pull/28) (@lox)
- Upload lambda to region specific buckets [#26](https://github.com/buildkite/buildkite-agent-metrics/pull/26) (@lox)

## [v1.4.1](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.4.1) (2016-12-20)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.4.0...v1.4.1)

### Changed
- Support the queue parameter and logs in lambda func [#25](https://github.com/buildkite/buildkite-agent-metrics/pull/25) (@lox)

## [v1.4.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.4.0) (2016-12-19)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.3.0...v1.4.0)

### Changed
- Add StatsD support [#24](https://github.com/buildkite/buildkite-agent-metrics/pull/24) (@callumj)

## [v1.3.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.3.0) (2016-12-19)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.2.0...v1.3.0)

### Changed
- Correctly filter stats by queue [#23](https://github.com/buildkite/buildkite-agent-metrics/pull/23) (@lox)
- Moved collector into subpackage with tests [#22](https://github.com/buildkite/buildkite-agent-metrics/pull/22) (@lox)
- Debug flag now shows useful debugging, added dry-run [#20](https://github.com/buildkite/buildkite-agent-metrics/pull/20) (@lox)
- Add a lambda function for executing stats [#18](https://github.com/buildkite/buildkite-agent-metrics/pull/18) (@lox)
- Add a quiet flag to close #9 [#14](https://github.com/buildkite/buildkite-agent-metrics/pull/14) (@lox)
- Revert "Support multiple queues via --queue" [#16](https://github.com/buildkite/buildkite-agent-metrics/pull/16) (@sj26)
- Support multiple queues via --queue [#13](https://github.com/buildkite/buildkite-agent-metrics/pull/13) (@lox)
- Increase page size [#12](https://github.com/buildkite/buildkite-agent-metrics/pull/12) (@lox)
- Replace glide with govendor, bump vendors [#11](https://github.com/buildkite/buildkite-agent-metrics/pull/11) (@lox)
- Improve error logging [#7](https://github.com/buildkite/buildkite-agent-metrics/pull/7) (@yeungda-rea)

## [v1.2.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.2.0) (2016-06-22)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.1.0...v1.2.0)

### Changed
- OpenJobsCount [#3](https://github.com/buildkite/buildkite-agent-metrics/pull/3) (@eliank)
- Add a -queue flag to allow filtering metrics by queue [#1](https://github.com/buildkite/buildkite-agent-metrics/pull/1) (@lox)

## [v1.1.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.1.0) (2016-04-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/v1.0.0...v1.1.0)

### Changed

## [v1.0.0](https://github.com/buildkite/buildkite-agent-metrics/tree/v1.0.0) (2016-04-07)
[Full Changelog](https://github.com/buildkite/buildkite-agent-metrics/compare/78a3ded05dcf...v1.0.0)

Initial release
