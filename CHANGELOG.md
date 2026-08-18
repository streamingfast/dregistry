# Change log

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

* Add public `sf.registry.v1.FoundationStoreRegistryService` proto and generated Go bindings for foundational-store endpoint lookup (`GetFoundationStore`).

* Add GitHub Actions for pull-request required checks, Buf lint/breaking checks, BSR publish, and tag-based GitHub releases (changelog extracted with `sfreleaser`).
