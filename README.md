# buf-lint-extra

Extra lint rules for [buf](https://buf.build), packaged as a
[buf check plugin](https://buf.build/docs/cli/buf-plugins/overview/).

| Rule                            | Default | What it checks                                                                                                                                                        |
| ------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ENUM_DEDICATED_FILE`           | on      | Each enum lives at the top level of a file of its own: no other enums, messages, services, or extensions, and no enums nested in messages.                            |
| `ENUM_DEDICATED_PACKAGE`        | off     | Files that declare only enums have a package whose last component is the name of the enum, in any case style, so every enum gets a package of its own.                |
| `ENUM_FILE_MATCH`               | off     | Files that declare only enums are named after the enum, in any case style, with or without the enum file suffix.                                                      |
| `ENUM_FILE_SUFFIX`              | off     | Files that declare top-level enums have a name ending in a suffix (`_enum` by default), and files with that suffix declare top-level enums.                           |
| `FILE_LOWER_KEBAB_CASE`         | off     | File names are lower-kebab-case in each dot-separated segment, such as `user-service.proto` or `user-status.enum.proto`.                                              |
| `PACKAGE_CAMEL_CASE`            | off     | Package names are camelCase: every dot-separated component starts with a lowercase letter and contains no underscores.                                                |
| `PACKAGE_DIRECTORY_MATCH_EXTRA` | off     | Files are in the directory matching their package, like the builtin rule, with options to exclude prefixes, convert the case, and leave out the enum's own component. |

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
    - ENUM_DEDICATED_PACKAGE
    - ENUM_FILE_MATCH
    - ENUM_FILE_SUFFIX
    - FILE_LOWER_KEBAB_CASE
    - PACKAGE_CAMEL_CASE
    - PACKAGE_DIRECTORY_MATCH_EXTRA
  except:
    - FILE_LOWER_SNAKE_CASE # part of STANDARD, contradicts FILE_LOWER_KEBAB_CASE
    - PACKAGE_LOWER_SNAKE_CASE # part of STANDARD, contradicts PACKAGE_CAMEL_CASE
    - PACKAGE_DIRECTORY_MATCH # part of STANDARD, replaced by PACKAGE_DIRECTORY_MATCH_EXTRA
plugins:
  - plugin: buf-plugin-lint-extra
    options:
      enum_file_suffix: -enum # optional, the default is _enum
      package_directory_excluded_prefixes: # optional, empty by default
        - acme.v1
      package_directory_case: lower-kebab-case # optional, components are used as is by default
      package_directory_enum_component: excluded # optional, excluded when ENUM_DEDICATED_PACKAGE is enabled, included otherwise
```

When `lint.use` is set, only the listed rules and categories run, so plugin
rules must be listed explicitly. When `lint.use` is omitted, `buf` runs its
`STANDARD` category plus the plugin's default rules, which is only
`ENUM_DEDICATED_FILE`.

Everything else works as for the builtin rules: `except`, `ignore`,
`ignore_only`, and `// buf:lint:ignore` comments.

## Rules

### ENUM_DEDICATED_FILE

Reports every enum that is not the only declaration of its file:

- a top-level enum in a file that also declares other enums, messages,
  services, or extensions;
- an enum nested in a message, at any depth.

Nested enums are reported because code generators emit them inside the module
of their message. ts-proto, for example, generates `User.Role` as `User_Role`
in the module of `User`, so code that needs only the enum has to import every
message of that file. Two enums in one file are reported for the same reason:
importing one of them brings in the other. A file with only file options,
imports, and a single enum is fine.

```proto
// user.proto: flagged, Status and Role have to move to files of their own
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
// status_enum.proto: ok, the file declares a single enum
enum Status {
  STATUS_UNSPECIFIED = 0;
}
```

```
acme/v1/user.proto:5:1:Enum "Status" must be declared in a file of its own, but this file also declares 1 message.
acme/v1/user.proto:10:3:Enum "User.Role" must be declared at the top level of a file of its own, not nested in message "User".
acme/v1/enums.proto:5:1:Enum "Status" must be declared in a file of its own, but this file also declares 1 other enum.
```

The annotation is attached to the enum, so a single enum, top-level or nested,
can be exempted with `// buf:lint:ignore ENUM_DEDICATED_FILE` on the line above
it.

### ENUM_DEDICATED_PACKAGE

Checks that a file declaring only enums has a package whose last component is
the name of the enum, so that every enum gets a package of its own, named
after it. Such a package keeps the generated code of the enum apart from
everything else, in the same way `ENUM_DEDICATED_FILE` keeps the enum apart
in the sources. With two enums in one file, at most one can match, so the
other is reported and has to move to its own package and file.

The comparison looks at the words of the names and ignores their case style:
enum `OrderStatus` matches both `app.test.orderStatus` and
`app.test.order_status`, but not `app.test.orderstatus`. The case style of the
package is left to `PACKAGE_CAMEL_CASE` or the builtin
`PACKAGE_LOWER_SNAKE_CASE`, so the two rules never ask for different things.
Files that also declare messages, services, or extensions, and files without
a package, are not checked.

```proto
// kind.proto: flagged, the package is not named after the enum
package app.test.enums;

enum Kind {
  KIND_UNSPECIFIED = 0;
}
```

```
app/test/kind.proto:5:1:Enum "Kind" must be declared in a package named after it, such as "app.test.kind", but the package is "app.test.enums".
```

When the camelCase and lower_snake_case forms of the name differ, both are
suggested, unless `PACKAGE_CAMEL_CASE` is enabled, in which case only the
camelCase form is. The annotation is attached to the enum, so it can be
exempted with `// buf:lint:ignore ENUM_DEDICATED_PACKAGE` on the line above
it.

Such packages add a directory per enum under the builtin
`PACKAGE_DIRECTORY_MATCH`. `PACKAGE_DIRECTORY_MATCH_EXTRA` leaves that
component out of the directory whenever this rule is enabled, so the enum
files stay next to the files of the parent package; see its
`package_directory_enum_component` option.

The rule is off by default. Enable it by listing `ENUM_DEDICATED_PACKAGE` in
`lint.use`.

### ENUM_FILE_MATCH

Checks that a file declaring only enums is named after the enum, so that the
enum can be found by its file name. This matters most with
`ENUM_DEDICATED_PACKAGE` and `PACKAGE_DIRECTORY_MATCH_EXTRA`, where the
directory no longer carries the name of the enum: with all three, the file
`app/test/dedicated.proto`, the package `app.test.dedicated` and the enum
`Dedicated` all agree. With two enums in one file, at most one can match, so
the other is reported.

The name is compared without the `.proto` extension and, when present,
without the `enum_file_suffix` of `ENUM_FILE_SUFFIX`, `_enum` by default. As
with `ENUM_DEDICATED_PACKAGE`, the comparison looks at the words and ignores
their case style: enum `OrderStatus` matches `order-status.proto`,
`order_status.proto` and `order-status_enum.proto`, but not
`orderstatus.proto`. The case style of the file name is left to
`FILE_LOWER_KEBAB_CASE` or the builtin `FILE_LOWER_SNAKE_CASE`. Files that
also declare messages, services, or extensions are not checked.

```
acme/v1/misc-enums.proto:5:1:Enum "Severity" must be declared in a file named after it, such as "acme/v1/severity-enums.proto", but the file is "acme/v1/misc-enums.proto".
```

The suggestion has the suffix when the file has one or `ENUM_FILE_SUFFIX` is
enabled, and lists both the lower-kebab-case and the lower_snake_case form
when they differ, unless `FILE_LOWER_KEBAB_CASE` is enabled, in which case
only the lower-kebab-case form is. The annotation is attached to the enum, so
it can be exempted with `// buf:lint:ignore ENUM_FILE_MATCH` on the line
above it.

The rule is off by default. Enable it by listing `ENUM_FILE_MATCH` in
`lint.use`.

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

### FILE_LOWER_KEBAB_CASE

Checks that the file name, without its directories and the `.proto` extension,
is lower-kebab-case: lowercase words separated by single hyphens. This is the
kebab-case counterpart of the builtin `FILE_LOWER_SNAKE_CASE`, which wants
`user_service.proto` where this rule wants `user-service.proto`. Words are
split at underscores and at capital letters, so `UserService.proto` and
`userService.proto` are flagged too, with `user-service.proto` as the
suggestion.

Dots separate segments, and each segment is checked on its own, so a dotted
suffix such as `.enum.proto` is fine: `user-status.enum.proto` passes, and
`User_Status.enum.proto` is flagged with `user-status.enum.proto` as the
suggestion. The builtin rule differs here, it treats dots as word separators.

```
acme/v1/order_item.proto:1:1:Filename "order_item.proto" should be lower-kebab-case.proto, such as "order-item.proto".
acme/v1/User_Status.enum.proto:1:1:Filename "User_Status.enum.proto" should be lower-kebab-case.proto, such as "user-status.enum.proto".
```

Annotations are reported at the file level, because the fix is to rename the
file. Use `lint.ignore` or `lint.ignore_only` in `buf.yaml` to exclude paths.

The rule contradicts `FILE_LOWER_SNAKE_CASE`, which is part of the `STANDARD`
category, so list that rule in `lint.except` when using it. Directories are
not checked, so `PACKAGE_DIRECTORY_MATCH` is unaffected; to have kebab-case
directories as well, see `PACKAGE_DIRECTORY_MATCH_EXTRA`. When combined with
`ENUM_FILE_SUFFIX`, set `enum_file_suffix` to a kebab-case suffix such as
`-enum`, or to a dotted one such as `.enum`, since the default `_enum` would
be flagged.

The rule is off by default. Enable it by listing `FILE_LOWER_KEBAB_CASE` in
`lint.use`.

### PACKAGE_CAMEL_CASE

Checks that every dot-separated component of the package name is camelCase: it
starts with a lowercase letter and contains no underscores. This is the
camelCase counterpart of the builtin `PACKAGE_LOWER_SNAKE_CASE`, which wants
`acme.user_service.v1` where this rule wants `acme.userService.v1`.

Uppercase letters after the first one are accepted, so `acme.userAPI.v1` is
fine, in the same way that the builtin `PASCAL_CASE` rules accept `HTTPServer`.
Version components such as `v1` and `v1beta1` are camelCase already.

```proto
// flagged
package acme.user_service.v1;
```

```proto
// ok
package acme.userService.v1;
```

```
acme/user_service/v1/profile.proto:3:1:Package name "acme.user_service.v1" should be camelCase, such as "acme.userService.v1".
```

The annotation is attached to the `package` statement, so a file can be
exempted with `// buf:lint:ignore PACKAGE_CAMEL_CASE` on the line above it.

The rule contradicts `PACKAGE_LOWER_SNAKE_CASE`, which is part of the `BASIC`
and `STANDARD` categories, so list that rule in `lint.except` when using one of
them. `PACKAGE_DIRECTORY_MATCH` is unaffected and keeps asking for the
directory to match the package, `acme/userService/v1/` in the example above;
`PACKAGE_DIRECTORY_MATCH_EXTRA` can map that to `acme/user-service/v1/`
instead. `FILE_LOWER_SNAKE_CASE` only looks at the file name, not at its
directories.

The rule is off by default. Enable it by listing `PACKAGE_CAMEL_CASE` in
`lint.use`.

### PACKAGE_DIRECTORY_MATCH_EXTRA

Checks that a file is in the directory, relative to the module root, that
matches its package, with one directory per dot-separated component. Without
options it behaves like the builtin `PACKAGE_DIRECTORY_MATCH`: files with
package `acme.billing.v1` must be in `acme/billing/v1/`. Two options change
how the expected directory is derived from the package.

`package_directory_excluded_prefixes` lists package prefixes that are not
expected in the directory. A prefix matches whole components, and when several
match, the longest wins. With `acme.v1` excluded, files with package
`acme.v1.billing` are checked as if their package were `billing`, so they must
be in `billing/`. Files whose package is exactly an excluded prefix are not
checked, like files without a package.

`package_directory_case` converts each component to the case directories use,
either `lower-kebab-case` or `lower_snake_case`. With `lower-kebab-case`, files
with package `acme.mainGoal.v1` must be in `acme/main-goal/v1/`, and
`acme/mainGoal/v1/` is reported, because the conversion is exact rather than
case-insensitive. Components that are already in that case, such as `acme`
and `v1`, are unchanged.

`package_directory_enum_component` decides what happens to the package
component named after the enum, the one `ENUM_DEDICATED_PACKAGE` asks for.
With `excluded`, a file that declares only enums has the last component of its
package dropped when it is the name of one of those enums in any case style,
so an enum file `Dedicated` with package `app.test.dedicated` must be in
`app/test/`, next to the files of `app.test`, and `app/test/dedicated/` is
reported. Enum files whose last component is not an enum name, and files that
also declare messages, services, or extensions, keep the full package. With
`included`, the component is a directory like any other. When the option is
not set, the component is `excluded` if `ENUM_DEDICATED_PACKAGE` is enabled,
that is listed in `lint.use` and not in `lint.except`, and `included`
otherwise, so the two rules agree on the layout unless told otherwise.

The excluded prefix is removed first, then the enum component, and the rest is
converted. If nothing is left, the file is not checked, like a file without a
package.

```yaml
plugins:
  - plugin: buf-plugin-lint-extra
    options:
      package_directory_excluded_prefixes:
        - legacy
      package_directory_case: lower-kebab-case
      package_directory_enum_component: excluded
```

With this configuration, files with package `legacy.acme.billing.v2` must be
in `acme/billing/v2/`, and an enum file declaring `Priority` with package
`acme.v1.priority` must be in `acme/v1/`:

```
acme/billing/receipt.proto:3:1:Files with package "legacy.acme.billing.v2" must be within a directory "acme/billing/v2" relative to root but were in directory "acme/billing".
```

The annotation is attached to the `package` statement, so a file can be
exempted with `// buf:lint:ignore PACKAGE_DIRECTORY_MATCH_EXTRA` on the line
above it.

The rule replaces `PACKAGE_DIRECTORY_MATCH`, which is part of the `MINIMAL`,
`BASIC` and `STANDARD` categories, so list that rule in `lint.except` when
using one of them, or both rules will check the same files with different
expectations.

The rule is off by default. Enable it by listing
`PACKAGE_DIRECTORY_MATCH_EXTRA` in `lint.use`.

## Development

```sh
go test ./...
go build ./cmd/buf-plugin-lint-extra
GOOS=wasip1 GOARCH=wasm go build -o buf-plugin-lint-extra.wasm ./cmd/buf-plugin-lint-extra
```

Rules live in [internal/rules](internal/rules). Each rule is a function from
a `fileSummary`, the few facts about a file that the rules look at, to
annotations. It is registered as a `check.RuleSpec` of
[bufplugin-go](https://github.com/bufbuild/bufplugin-go) and covered by
`checktest` tests that compile the `.proto` fixtures under
[internal/rules/testdata](internal/rules/testdata) and compare the produced
annotations with the expected ones.

The plugin does not serve `check` through bufplugin-go, though.
[server.go](internal/rules/server.go) runs its own pluginrpc server, and
[request.go](internal/rules/request.go) builds the summaries with a
`protowire` scan of the request instead of unmarshaling and validating the
descriptors: `buf` starts the plugin several times per lint run, and under
Wasm that work took seconds on larger modules. `ListRules`, `ListCategories`
and `GetPluginInfo` still use the bufplugin-go handlers. The `check.RuleSpec`
handlers feed the same parser from bufplugin-go's descriptors, and
`TestServerMatchesSpec` runs every fixture through both servers, so the two
cannot drift.

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
