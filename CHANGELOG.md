# Changelog

All notable changes to Chaosd will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [1.5.0] - 2026-10-20

### Changed

- Upgrade Go to 1.27.2 and refresh core runtime dependencies and bundled fault injection tools. ([#277](https://github.com/chaos-mesh/chaosd/pull/277), [#278](https://github.com/chaos-mesh/chaosd/pull/278))
- Make Docker builds self-contained and include the required runtime utilities and helper tools. Newly built release archives require glibc 2.34 or later. ([#281](https://github.com/chaos-mesh/chaosd/pull/281), [#284](https://github.com/chaos-mesh/chaosd/pull/284))

### Fixed

- Restore HTTP fault creation through the HTTP API. ([#280](https://github.com/chaos-mesh/chaosd/pull/280))
- Honor disk read and write durations instead of allowing workloads to continue past their deadlines. ([#280](https://github.com/chaos-mesh/chaosd/pull/280))
- Correct experiment history ordering, support scheduled experiment searches, and fix SQLite schema index conflicts. ([#279](https://github.com/chaos-mesh/chaosd/pull/279))
- Accept valid Redis key expiration options and percentage-based cache limits. ([#280](https://github.com/chaos-mesh/chaosd/pull/280))
- Report recovery failures for non-recoverable faults instead of returning success. ([#280](https://github.com/chaos-mesh/chaosd/pull/280))

[Unreleased]: https://github.com/chaos-mesh/chaosd/compare/v1.5.0...HEAD
[1.5.0]: https://github.com/chaos-mesh/chaosd/compare/v1.4.1...v1.5.0
