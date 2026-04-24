# Changelog

The list of commits in this changelog is automatically generated in the release process.
The commits follow the Conventional Commit specification.

## [0.2.1] - 2026-04-24

### 🚀 Features

- Implemented git sha and git tag injection during build time (#45)
- Log profile used for signing certificate in Run method (#41)
- Use hash tagged actions, fix lint issues (#38)
- Updated go deps & updated minimal required go version (#42)
- Add workflow lint (#39)

### 🐛 Bug Fixes

- Adjust permissions for docker release, add GitHub release stage (#48)
- Adjust workflow permissions (#46)
- Refactor workflow lint action (#43)
- Adjust permissions for nightly workflow (#44)

### 💼 Other

- Added golangci lint  (#40)

### 🚜 Refactor

- Replace marshalled JSON responses with direct response logging in command handlers (#37)

### ⚙️ Miscellaneous Tasks

- Update actions to latest versions for Node 24 support (#36)

## [0.1.1-rc1] - 2026-03-26

### 🚀 Features

- Add new release workflow (#35)

## [0.1.1-rc0] - 2026-03-23

### 🚀 Features

- Add nightly security scan (#33)
- Updated all direct and transitive dependencies (#34)
- Updated test cases updated lib (#32)
- Updated go version to latest (#29)
- Log health check status in response (#27)
- Updated workflow so it displays details of server binary (#24)
- Updated crypto-broker-client-go lib version (#25)
- Updated Taskfile & created .env.example file (#23)
- Added step that runs benchmarks with server in FIPS mode (#21)
- Implemented e2e benchmarks, updated commands, linted code, defi… (#19)
- Updated library to latest version (#18)
- Updated OTEL package, updated Taskfile, introduced new dep (#17)
- Updated crypto-broker-client-go reference, created fake endpoint logic (#15)
- Add workflow for generating binary during release (#9)
- Changed RunE -> Run & PreRunE -> PreRun (#7)
- Add health check test to Taskfile (#10)
- Updated crypto-broker-client-go lib & adjusted code (#6)
- Add health command to check broker server status (#4)
- Updated crypto-broker-client-go library (#5)

### 🐛 Bug Fixes

- Update Dockerfile (#31)
- Adjust binary name (#30)
- Adjust env file and Taskfile (#28)

### 💼 Other

- Service version & name fix (#26)
- Updated Taskfile by adding vars related to OTEL (#16)
- Introduced otel tracing in cli (#13)

### 🚜 Refactor

- Adjust Docker image generation (#12)
- Adjust Task setup (#8)

## [0.1.0] - 2025-12-02

### 🚀 Features

- Adjust workflow files and remove local config files  (#1)
- Changed to new workflow for ghcr upload (#2)
- Initial commit

### 🐛 Bug Fixes

- Adjust task test-sign to new path (#3)
- Updated .gitignore
