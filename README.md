# buf-lint-extra

Extra lint rules for [buf](https://buf.build), packaged as a
[buf check plugin](https://buf.build/docs/cli/buf-plugins/overview/).

| Rule | Default | What it checks |
| --- | --- | --- |
| `ENUM_DEDICATED_FILE` | on | Top-level enums live in files that declare nothing but enums: no messages, services, or extensions. |
| `ENUM_FILE_SUFFIX` | off | Files that declare top-level enums have a name ending in a suffix (`_enum` by default), and files with that suffix declare top-level enums. |

## Installation

Either install the native binary with Go:

```sh
go install github.com/infodusha/buf-lint-extra/cmd/buf-plugin-lint-extra@latest
```

and reference it by name, in which case `buf-plugin-lint-extra` must be on your
`PATH` when `buf` runs:

```yaml
plugins:
  - plugin: buf-plugin-lint-extra
```

Or download `buf-plugin-lint-extra.wasm` from the
[latest release](https://github.com/infodusha/buf-lint-extra/releases/latest)
and reference the file by path. No Go toolchain is needed, `buf` runs the
WebAssembly module itself. The `.wasm` extension is required, it is how `buf`
tells a Wasm plugin from a native binary:

```yaml
plugins:
  - plugin: ./buf-plugin-lint-extra.wasm
```

## Configuration

```yaml
# buf.yaml
version: v2
lint:
  use:
    - STANDARD # omit if you do not want to use the rules builtin to buf
    - ENUM_DEDICATED_FILE
    - ENUM_FILE_SUFFIX
plugins:
  - plugin: buf-plugin-lint-extra
    options:
      enum_file_suffix: _enum # optional, this is the default
```

When `lint.use` is set, only the listed rules and categories run, so plugin
rules must be listed explicitly. When `lint.use` is omitted, `buf` runs its
`STANDARD` category plus the plugin's default rules, which is only
`ENUM_DEDICATED_FILE`.

Everything else works as for the builtin rules: `except`, `ignore`,
`ignore_only`, and `// buf:lint:ignore` comments.

## Rules

### ENUM_DEDICATED_FILE

Reports every top-level enum declared in a file that also declares messages,
services, or extensions. Enums nested inside messages belong to their message
and are not considered, so a message file with nested enums is fine, and a
file with only file options, imports, and enums is fine too.

```proto
// user.proto: flagged, Status has to move to its own file
enum Status {
  STATUS_UNSPECIFIED = 0;
}

message User {
  Status status = 1;
}
```

```proto
// status_enum.proto: ok, the file declares only enums
enum Status {
  STATUS_UNSPECIFIED = 0;
}

enum Role {
  ROLE_UNSPECIFIED = 0;
}
```

```
acme/v1/user.proto:5:1:Enum "Status" must be declared in a dedicated file that contains only enums, but this file also declares 1 message.
```

The annotation is attached to the enum, so a single enum can be exempted with
`// buf:lint:ignore ENUM_DEDICATED_FILE` on the line above it.

### ENUM_FILE_SUFFIX

Checks the relation between a file's name and whether it declares top-level
enums, in both directions:

- a file that declares top-level enums must have a name ending in the suffix;
- a file whose name ends in the suffix must declare top-level enums.

The suffix is `_enum` by default, so enum files look like `status_enum.proto`.
Override it with the `enum_file_suffix` option. The value may be given with or
without the `.proto` extension: `_enums` and `_enums.proto` mean the same thing.

```
acme/v1/status.proto:1:1:File "acme/v1/status.proto" declares top-level enums and must have a name ending in "_enum.proto", such as "acme/v1/status_enum.proto".
acme/v1/color_enum.proto:1:1:File "acme/v1/color_enum.proto" has a name ending in "_enum.proto" but declares no top-level enums.
```

Annotations are reported at the file level, because the fix is to rename the
file. Use `lint.ignore` or `lint.ignore_only` in `buf.yaml` to exclude paths.

The rule is off by default. Enable it by listing `ENUM_FILE_SUFFIX` in
`lint.use`.

## Development

```sh
go test ./...
go build ./cmd/buf-plugin-lint-extra
GOOS=wasip1 GOARCH=wasm go build -o buf-plugin-lint-extra.wasm ./cmd/buf-plugin-lint-extra
```

Rules live in [internal/rules](internal/rules). Each rule is a
`check.RuleSpec` with a handler built on
[bufplugin-go](https://github.com/bufbuild/bufplugin-go), and is covered by
`checktest` tests that compile the `.proto` fixtures under
[internal/rules/testdata](internal/rules/testdata) and compare the produced
annotations with the expected ones.

## Releases

Releases are managed by
[release-please](https://github.com/googleapis/release-please). Commits to
`main` follow [Conventional Commits](https://www.conventionalcommits.org/);
release-please keeps a release pull request up to date with the next version
and changelog. Merging that pull request tags the release and publishes a
GitHub release with `buf-plugin-lint-extra.wasm` and its SHA-256 checksum
attached. Versions are `0.x` until the rules are declared stable, so breaking
changes bump the minor version.

## License

[Apache 2.0](LICENSE)
