# buf-lint-extra

Extra lint rules for [buf](https://buf.build), packaged as a
[buf check plugin](https://buf.build/docs/cli/buf-plugins/overview/).

| Rule                  | Default | What it checks                                                                                                                              |
| --------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `ENUM_DEDICATED_FILE` | on      | Enums live at the top level of files that declare nothing but enums: no messages, services, or extensions, and no enums nested in messages. |
| `ENUM_FILE_SUFFIX`    | off     | Files that declare top-level enums have a name ending in a suffix (`_enum` by default), and files with that suffix declare top-level enums. |

## Installation

A native binary is the fastest option. `buf` runs the Wasm module in a
runtime, which adds a few hundred milliseconds to every lint run, and several
seconds to the first one while the module is compiled.

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

Or, on Linux, download the static binary for your architecture,
`buf-plugin-lint-extra-Linux-x86_64` or `buf-plugin-lint-extra-Linux-aarch64`,
from the [latest release](https://github.com/infodusha/buf-lint-extra/releases/latest):

```sh
curl -fsSL -o buf-plugin-lint-extra \
  "https://github.com/infodusha/buf-lint-extra/releases/latest/download/buf-plugin-lint-extra-$(uname -s)-$(uname -m)"
chmod +x buf-plugin-lint-extra
```

and reference it by name from your `PATH` as above, or by path, relative to the
directory `buf` runs in:

```yaml
plugins:
  - plugin: ./buf-plugin-lint-extra
```

Or download `buf-plugin-lint-extra.wasm` from the
[latest release](https://github.com/infodusha/buf-lint-extra/releases/latest)
and reference the file by path, relative to the directory `buf` runs in. No Go
toolchain is needed, `buf` runs the WebAssembly module itself. The `.wasm`
extension is required, it is how `buf` tells a Wasm plugin from a native
binary:

```yaml
plugins:
  - plugin: ./buf-plugin-lint-extra.wasm
```

Each release attaches a SHA-256 checksum and a signed build provenance
attestation for every binary and the module, which can be checked with the
[GitHub CLI](https://cli.github.com):

```sh
sha256sum --check buf-plugin-lint-extra.wasm.sha256
gh attestation verify buf-plugin-lint-extra.wasm --repo infodusha/buf-lint-extra
```

Or reference the same module from the
[Buf Schema Registry](https://buf.build/dusha/lint-extra), so that `buf`
downloads it and nothing has to be installed:

```yaml
plugins:
  - plugin: buf.build/dusha/lint-extra
```

and run `buf plugin update` to pin the plugin version in `buf.lock`.

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

Reports every enum that is not declared at the top level of a file containing
only enums:

- a top-level enum in a file that also declares messages, services, or
  extensions;
- an enum nested in a message, at any depth.

Nested enums are reported because code generators emit them inside the module
of their message. ts-proto, for example, generates `User.Role` as `User_Role`
in the module of `User`, so code that needs only the enum has to import every
message of that file. A file with only file options, imports, and enums is
fine.

```proto
// user.proto: flagged, Status and Role have to move to their own file
enum Status {
  STATUS_UNSPECIFIED = 0;
}

message User {
  enum Role {
    ROLE_UNSPECIFIED = 0;
  }

  Status status = 1;
  Role role = 2;
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
acme/v1/user.proto:10:3:Enum "User.Role" must be declared at the top level of a dedicated file that contains only enums, not nested in message "User".
```

The annotation is attached to the enum, so a single enum, top-level or nested,
can be exempted with `// buf:lint:ignore ENUM_DEDICATED_FILE` on the line above
it.

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

When the file also declares messages, services, or extensions, renaming it
would not help, so the rule asks for the enums to be moved out instead, in line
with `ENUM_DEDICATED_FILE`:

```
acme/v1/user.proto:1:1:File "acme/v1/user.proto" declares top-level enums alongside 1 message, so the enums must move to a file with a name ending in "_enum.proto".
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

The end-to-end tests in [e2e](e2e) build the plugin both as a native binary and
as a Wasm module, run `buf lint` with it on the module under
[e2e/testdata](e2e/testdata), and check the reported annotations, including
plugin options and both ways of ignoring them. They need `buf` on `PATH` and
are skipped without it, except in CI, and with `go test -short`.

## Releases

Releases are managed by
[release-please](https://github.com/googleapis/release-please). Commits to
`main` follow [Conventional Commits](https://www.conventionalcommits.org/);
release-please keeps a release pull request up to date with the next version
and changelog. Merging that pull request tags the release and publishes a
GitHub release with `buf-plugin-lint-extra.wasm`, its SHA-256 checksum, and a
build provenance attestation. The same module is then pushed to
[buf.build/dusha/lint-extra](https://buf.build/dusha/lint-extra),
labelled with both `main` and the release tag. That push needs a BSR token in
the `BUF_TOKEN` repository secret. Versions are `0.x` until the rules are declared
stable, so breaking changes bump the minor version.

## License

[Apache 2.0](LICENSE)
