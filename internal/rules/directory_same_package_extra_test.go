package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

const directorySamePackageExtraTestdataDir = "testdata/directory_same_package_extra"

type protoFile struct {
	name string
	pkg  string
}

func TestDirectorySamePackageExtra(t *testing.T) {
	t.Parallel()

	excluded := map[string]any{PackageDirectoryEnumComponentOptionKey: "excluded"}

	for _, testCase := range []struct {
		name    string
		files   []protoFile
		options map[string]any
		ruleIDs []string
		message string
	}{
		{
			name:  "files of one package are accepted",
			files: []protoFile{{"acme/v1/order.proto", "acme.v1"}, {"acme/v1/user.proto", "acme.v1"}},
		},
		{
			name:    "every file of a directory with two packages is flagged",
			files:   []protoFile{{"acme/v1/order.proto", "acme.v1"}, {"acme/v1/stray.proto", "acme.v2"}, {"acme/v1/user.proto", "acme.v1"}},
			message: `Multiple packages "acme.v1,acme.v2" detected within directory "acme/v1".`,
		},
		{
			name:    "file without a package is flagged with the package",
			files:   []protoFile{{"acme/v1/no_package.proto", ""}, {"acme/v1/user.proto", "acme.v1"}},
			message: `Package "acme.v1" and file with no package detected within directory "acme/v1".`,
		},
		{
			name:    "file without a package is flagged with the packages",
			files:   []protoFile{{"acme/v1/no_package.proto", ""}, {"acme/v1/stray.proto", "acme.v2"}, {"acme/v1/user.proto", "acme.v1"}},
			message: `Multiple packages "acme.v1,acme.v2" and file with no package detected within directory "acme/v1".`,
		},
		{
			name:    "files in the root directory are checked",
			files:   []protoFile{{"other.proto", "other"}, {"root.proto", "acme.v1"}},
			message: `Multiple packages "acme.v1,other" detected within directory ".".`,
		},
		{
			name:  "files in different directories are not compared",
			files: []protoFile{{"acme/v1/user.proto", "acme.v1"}, {"app/test/holder.proto", "app.test"}},
		},
		{
			name:    "enum component is a package of its own by default",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			message: `Multiple packages "app.test,app.test.dedicated" detected within directory "app/test".`,
		},
		{
			name:    "enum file counts as the parent package when the enum component is excluded",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			options: excluded,
		},
		{
			name:    "enum component is excluded by default when ENUM_DEDICATED_PACKAGE runs",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			ruleIDs: []string{DirectorySamePackageExtraRuleID, EnumDedicatedPackageRuleID},
		},
		{
			name:    "included keeps the enum component when ENUM_DEDICATED_PACKAGE runs",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			ruleIDs: []string{DirectorySamePackageExtraRuleID, EnumDedicatedPackageRuleID},
			options: map[string]any{PackageDirectoryEnumComponentOptionKey: "included"},
			message: `Multiple packages "app.test,app.test.dedicated" detected within directory "app/test".`,
		},
		{
			name:    "enum component matches the enum name in any case style",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/order-status.proto", "app.test.orderStatus"}},
			options: excluded,
		},
		{
			name:    "last component that is not the enum name is kept",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/kind.proto", "app.test.enums"}},
			options: excluded,
			message: `Multiple packages "app.test,app.test.enums" detected within directory "app/test".`,
		},
		{
			name:    "file with enums and messages keeps the enum component",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/mixed.proto", "app.test.mixed"}},
			options: excluded,
			message: `Multiple packages "app.test,app.test.mixed" detected within directory "app/test".`,
		},
		{
			name:    "enum file of another parent package is flagged with the parent",
			files:   []protoFile{{"app/test/foreign.proto", "app.other.foreign"}, {"app/test/holder.proto", "app.test"}},
			options: excluded,
			message: `Multiple packages "app.other,app.test" detected within directory "app/test".`,
		},
		{
			name:    "enum file whose package is only the enum component counts as a file without a package",
			files:   []protoFile{{"anywhere/loose.proto", ""}, {"anywhere/status.proto", "status"}},
			options: excluded,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ruleIDs := testCase.ruleIDs
			if ruleIDs == nil {
				ruleIDs = []string{DirectorySamePackageExtraRuleID}
			}
			checktest.CheckTest{
				Request: &checktest.RequestSpec{
					Files: &checktest.ProtoFileSpec{
						DirPaths:  []string{directorySamePackageExtraTestdataDir},
						FilePaths: protoFileNames(testCase.files),
					},
					RuleIDs: ruleIDs,
					Options: testCase.options,
				},
				Spec:                Spec,
				ExpectedAnnotations: directorySamePackageAnnotations(testCase.message, testCase.files),
			}.Run(t)
		})
	}
}

func TestDirectorySamePackageExtraInvalidOption(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	request, err := (&checktest.RequestSpec{
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{directorySamePackageExtraTestdataDir},
			FilePaths: []string{"acme/v1/user.proto"},
		},
		RuleIDs: []string{DirectorySamePackageExtraRuleID},
		Options: map[string]any{PackageDirectoryEnumComponentOptionKey: "dropped"},
	}).ToRequest(ctx)
	require.NoError(t, err)
	client, err := check.NewClientForSpec(Spec)
	require.NoError(t, err)
	_, err = client.Check(ctx, request)
	require.ErrorContains(t, err, PackageDirectoryEnumComponentOptionKey)
}

func protoFileNames(files []protoFile) []string {
	names := make([]string, len(files))
	for i, file := range files {
		names[i] = file.name
	}
	return names
}

func directorySamePackageAnnotations(message string, files []protoFile) []checktest.ExpectedAnnotation {
	if message == "" {
		return nil
	}
	annotations := make([]checktest.ExpectedAnnotation, len(files))
	for i, file := range files {
		location := &checktest.ExpectedFileLocation{FileName: file.name}
		if file.pkg != "" {
			location.StartLine, location.EndLine = 2, 2
			location.EndColumn = len("package " + file.pkg + ";")
		}
		annotations[i] = checktest.ExpectedAnnotation{
			RuleID:       DirectorySamePackageExtraRuleID,
			Message:      message,
			FileLocation: location,
		}
	}
	return annotations
}
