# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/).

> **Types of changes:**
>
> - **Added**: for new features.
> - **Changed**: for changes in existing functionality.
> - **Deprecated**: for soon-to-be removed features.
> - **Removed**: for now removed features.
> - **Fixed**: for any bug fixes.
> - **Security**: in case of vulnerabilities.

## [[UNRELEASED](https://github.com/sysflow-telemetry/sf-processor/compare/0.2.2...HEAD)]

## [[0.2.2](https://github.com/sysflow-telemetry/sf-processor/compare/0.2.1...0.2.2)] - 2020-12-07

### Changed

- Updated dependencies to latest `sf-apis`.

## [[0.2.1](https://github.com/sysflow-telemetry/sf-processor/compare/0.2.0...0.2.1)] - 2020-12-02

### Fixed

- Fixes `sf.file.oid` and `sf.file.newoid` attribute mapping.

## [[0.2.0](https://github.com/sysflow-telemetry/sf-processor/compare/0.1.0...0.2.0)] - 2020-12-01

### Added

- Adds lists and macro preprocessing to deal with usage before declarations in input policy language.
- Adds empty handling for process flow objects.
- Adds `endswith` binary operator to policy expression language.
- Added initial documentation.

### Changed

- Updates the grammar and intepreter to support falco policies.
- Several refactorings and performance optimizations in policy engine.
- Tuned filter policy for k8s clusters.

### Fixed

- Fixes module names and package paths.

## [0.1.0] - 2020-10-30

### Added

- First release of SysFlow Processor.
