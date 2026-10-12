package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

const directorySamePackageExtraTestdataDir = "testdata/directory_same_package_extra"

type protoFile struct {
	name string
	pkg  string
}

func TestDirectorySamePackageExtra(t *testing.T) {
	t.Parallel()

	withEnumDedicatedPackage := []string{DirectorySamePackageExtraRuleID, EnumDedicatedPackageRuleID}

	for _, testCase := range []struct {
		name    string
		files   []protoFile
		ruleIDs []string
		message string
		others  []checktest.ExpectedAnnotation
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
			name:    "enum file is a package of its own without ENUM_DEDICATED_PACKAGE",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			message: `Multiple packages "app.test,app.test.dedicated" detected within directory "app/test".`,
		},
		{
			name:    "enum file counts as the parent package with ENUM_DEDICATED_PACKAGE",
			files:   []protoFile{{"app/test/dedicated.proto", "app.test.dedicated"}, {"app/test/holder.proto", "app.test"}},
			ruleIDs: withEnumDedicatedPackage,
		},
		{
			name:    "enum component matches the enum name in any case style",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/order-status.proto", "app.test.orderStatus"}},
			ruleIDs: withEnumDedicatedPackage,
		},
		{
			name:    "last component that is not the enum name is kept",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/kind.proto", "app.test.enums"}},
			ruleIDs: withEnumDedicatedPackage,
			message: `Multiple packages "app.test,app.test.enums" detected within directory "app/test".`,
			others: []checktest.ExpectedAnnotation{{
				RuleID:       EnumDedicatedPackageRuleID,
				Message:      `Enum "Kind" must be declared in a package named after it, such as "app.test.kind", but the package is "app.test.enums".`,
				FileLocation: enumLocation("app/test/kind.proto", 0),
			}},
		},
		{
			name:    "file with enums and messages keeps the enum component",
			files:   []protoFile{{"app/test/holder.proto", "app.test"}, {"app/test/mixed.proto", "app.test.mixed"}},
			ruleIDs: withEnumDedicatedPackage,
			message: `Multiple packages "app.test,app.test.mixed" detected within directory "app/test".`,
		},
		{
			name:    "enum file of another parent package is flagged with the parent",
			files:   []protoFile{{"app/test/foreign.proto", "app.other.foreign"}, {"app/test/holder.proto", "app.test"}},
			ruleIDs: withEnumDedicatedPackage,
			message: `Multiple packages "app.other,app.test" detected within directory "app/test".`,
		},
		{
			name:    "enum file whose package is only the enum component counts as a file without a package",
			files:   []protoFile{{"anywhere/loose.proto", ""}, {"anywhere/status.proto", "status"}},
			ruleIDs: withEnumDedicatedPackage,
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
				},
				Spec:                Spec,
				ExpectedAnnotations: append(directorySamePackageAnnotations(testCase.message, testCase.files), testCase.others...),
			}.Run(t)
		})
	}
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
