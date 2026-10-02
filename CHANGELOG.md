# Changelog

## [0.3.1](https://github.com/infodusha/buf-lint-extra/compare/v0.3.0...v0.3.1) (2026-10-02)


### Bug Fixes

* trigger release ([7d34e7d](https://github.com/infodusha/buf-lint-extra/commit/7d34e7dacf9d3d0b95f142546d66d26546a59689))

## [0.3.0](https://github.com/infodusha/buf-lint-extra/compare/v0.2.1...v0.3.0) (2026-10-02)


### ⚠ BREAKING CHANGES

* ENUM_DEDICATED_FILE, which is on by default, now reports enums nested in messages. Move them to the top level of a file that declares only enums, or ignore them.

### Features

* report enums nested in messages in ENUM_DEDICATED_FILE ([f5a1bbf](https://github.com/infodusha/buf-lint-extra/commit/f5a1bbf167c73f776a4f32979e4839052d168640))


### Performance Improvements

* speed up the plugin under Wasm ([0b4ea1c](https://github.com/infodusha/buf-lint-extra/commit/0b4ea1c8928342c6e4334978ede170e56d296772))

## [0.2.1](https://github.com/infodusha/buf-lint-extra/compare/v0.2.0...v0.2.1) (2026-10-02)


### Bug Fixes

* **deps:** bump the go group with 2 updates ([#3](https://github.com/infodusha/buf-lint-extra/issues/3)) ([a4643d0](https://github.com/infodusha/buf-lint-extra/commit/a4643d08c28fa86102be37a1360a3a285ac60a09))

## [0.2.0](https://github.com/infodusha/buf-lint-extra/compare/v0.1.0...v0.2.0) (2026-10-02)


### Features

* ci updates + extra error ([362fcfc](https://github.com/infodusha/buf-lint-extra/commit/362fcfc0e98a8ddd97b90f2f536cb899bd0ad950))

## 0.1.0 (2026-10-01)


### Features

* add ENUM_DEDICATED_FILE and ENUM_FILE_SUFFIX lint rules ([eda4034](https://github.com/infodusha/buf-lint-extra/commit/eda4034438db85585efcdbf637cd3c51ca9d012f))
