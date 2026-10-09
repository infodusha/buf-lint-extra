# Changelog

## [0.5.0](https://github.com/infodusha/buf-lint-extra/compare/v0.4.0...v0.5.0) (2026-10-09)


### Features

* add ENUM_DEDICATED_PACKAGE rule and enum component option ([c54a1fb](https://github.com/infodusha/buf-lint-extra/commit/c54a1fbf9b053eb7dcc1e9b76e7ba2eaea3d15a1))

## [0.4.0](https://github.com/infodusha/buf-lint-extra/compare/v0.3.2...v0.4.0) (2026-10-09)


### Features

* add FILE_LOWER_KEBAB_CASE rule ([db8b510](https://github.com/infodusha/buf-lint-extra/commit/db8b510f176881736f5a51cfe763e4510946c32a))
* add PACKAGE_CAMEL_CASE rule ([387bccb](https://github.com/infodusha/buf-lint-extra/commit/387bccb55beee47374641f77c6ee2c8725af7c87))
* add PACKAGE_DIRECTORY_MATCH_EXTRA rule ([e34e779](https://github.com/infodusha/buf-lint-extra/commit/e34e7793155961925cf8e661cf9a6785c9e6d703))


### Bug Fixes

* **deps:** bump the go group with 2 updates ([c74032f](https://github.com/infodusha/buf-lint-extra/commit/c74032f3e422118dbc2d65bad073933f77f09fb4))
* **deps:** bump the go group with 2 updates ([#9](https://github.com/infodusha/buf-lint-extra/issues/9)) ([644ecf6](https://github.com/infodusha/buf-lint-extra/commit/644ecf60c2ea1a5ffb97247c6f9adc37600cd526))

## [0.3.2](https://github.com/infodusha/buf-lint-extra/compare/v0.3.1...v0.3.2) (2026-10-02)


### Bug Fixes

* **deps:** bump github.com/google/cel-go from 0.27.0 to 0.29.0 ([eca5d72](https://github.com/infodusha/buf-lint-extra/commit/eca5d72790ef64fa01a84538c3395c890f96143c))
* **deps:** bump github.com/google/cel-go from 0.27.0 to 0.29.0 ([#7](https://github.com/infodusha/buf-lint-extra/issues/7)) ([feb7fb6](https://github.com/infodusha/buf-lint-extra/commit/feb7fb6b3d0c9c43c428a1f66568d8ceac72ef6a))

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
