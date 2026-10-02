// Package e2e runs the plugin through the buf CLI, both as a native binary and
// as a Wasm module, to check that buf loads and configures it the way the
// README describes. The tests are skipped when buf is not on PATH, except in CI.
package e2e

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// bufYAML is formatted with the plugin reference. It covers plugin options,
// lint.ignore_only, and comment ignores, which buf applies, not the plugin.
const bufYAML = `version: v2
lint:
  use:
    - ENUM_DEDICATED_FILE
    - ENUM_FILE_SUFFIX
  ignore_only:
    ENUM_FILE_SUFFIX:
      - acme/v1/account.proto
plugins:
  - plugin: %s
    options:
      enum_file_suffix: _enums
`

// bufLintFailureExitCode is the exit code of buf lint when it reports annotations.
const bufLintFailureExitCode = 100

type annotation struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	Type        string `json:"type"`
	Message     string `json:"message"`
}

var expectedAnnotations = []annotation{
	{
		Path:        "acme/v1/color_enums.proto",
		StartLine:   1,
		StartColumn: 1,
		Type:        "ENUM_FILE_SUFFIX",
		Message:     `File "acme/v1/color_enums.proto" has a name ending in "_enums.proto" but declares no top-level enums.`,
	},
	{
		Path:        "acme/v1/kind.proto",
		StartLine:   1,
		StartColumn: 1,
		Type:        "ENUM_FILE_SUFFIX",
		Message:     `File "acme/v1/kind.proto" declares top-level enums and must have a name ending in "_enums.proto", such as "acme/v1/kind_enums.proto".`,
	},
	{
		Path:        "acme/v1/user.proto",
		StartLine:   1,
		StartColumn: 1,
		Type:        "ENUM_FILE_SUFFIX",
		Message:     `File "acme/v1/user.proto" declares top-level enums alongside 1 message, so the enums must move to a file with a name ending in "_enums.proto".`,
	},
	{
		Path:        "acme/v1/user.proto",
		StartLine:   5,
		StartColumn: 1,
		Type:        "ENUM_DEDICATED_FILE",
		Message:     `Enum "Status" must be declared in a dedicated file that contains only enums, but this file also declares 1 message.`,
	},
}

func TestBufLint(t *testing.T) {
	t.Parallel()

	buf := lookPathBuf(t)

	for _, testCase := range []struct {
		name  string
		goEnv []string
		// output is where the plugin is built, relative to the module directory.
		output string
		// plugin is the plugin reference in buf.yaml.
		plugin string
	}{
		{
			// Referenced by name, so its directory is put on PATH.
			name:   "native",
			output: "bin/buf-plugin-lint-extra",
			plugin: "buf-plugin-lint-extra",
		},
		{
			// Referenced by a path relative to the directory buf runs in.
			name:   "wasm",
			goEnv:  []string{"GOOS=wasip1", "GOARCH=wasm"},
			output: "buf-plugin-lint-extra.wasm",
			plugin: "./buf-plugin-lint-extra.wasm",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			moduleDir := t.TempDir()
			require.NoError(t, os.CopyFS(moduleDir, os.DirFS("testdata")))
			require.NoError(t, os.WriteFile(
				filepath.Join(moduleDir, "buf.yaml"),
				fmt.Appendf(nil, bufYAML, testCase.plugin),
				0o644,
			))
			goBuild(t, filepath.Join(moduleDir, testCase.output), testCase.goEnv)

			cmd := exec.CommandContext(t.Context(), buf, "lint", "--error-format=json")
			cmd.Dir = moduleDir
			cmd.Env = append(
				os.Environ(),
				"PATH="+filepath.Join(moduleDir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
				"BUF_CACHE_DIR="+t.TempDir(),
			)
			stdout, err := cmd.Output()
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr, "buf lint must report annotations")
			require.Equal(t, bufLintFailureExitCode, exitErr.ExitCode(), "buf lint stderr:\n%s", exitErr.Stderr)
			require.Equal(t, expectedAnnotations, parseAnnotations(t, stdout))
		})
	}
}

func lookPathBuf(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping buf end-to-end test in short mode")
	}
	buf, err := exec.LookPath("buf")
	if err != nil {
		// CI installs buf, so a missing buf there is a broken setup, not a reason to skip.
		if os.Getenv("CI") != "" {
			t.Fatalf("buf is required in CI: %v", err)
		}
		t.Skipf("buf is not on PATH: %v", err)
	}
	return buf
}

func goBuild(t *testing.T, output string, env []string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", output, "github.com/infodusha/buf-lint-extra/cmd/buf-plugin-lint-extra")
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go build:\n%s", out)
}

func parseAnnotations(t *testing.T, output []byte) []annotation {
	t.Helper()
	var annotations []annotation
	for line := range bytes.Lines(bytes.TrimSpace(output)) {
		var a annotation
		require.NoError(t, json.Unmarshal(line, &a), "buf lint output line: %s", line)
		annotations = append(annotations, a)
	}
	slices.SortFunc(annotations, func(a, b annotation) int {
		return cmp.Or(
			cmp.Compare(a.Path, b.Path),
			cmp.Compare(a.StartLine, b.StartLine),
			cmp.Compare(a.Type, b.Type),
		)
	})
	return annotations
}
