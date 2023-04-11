# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

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
