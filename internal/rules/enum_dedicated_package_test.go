package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

const enumDedicatedPackageTestdataDir = "testdata/enum_dedicated_package"

func enumLocation(file string, index int) *checktest.ExpectedFileLocation {
	return &checktest.ExpectedFileLocation{
		FileName:    file,
		StartLine:   4 + 4*index,
		StartColumn: 0,
		EndLine:     6 + 4*index,
		EndColumn:   1,
	}
}

func TestEnumDedicatedPackage(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name                string
		file                string
		ruleIDs             []string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "last component equal to the enum name is accepted",
			file: "status.proto",
		},
		{
			name: "camelCase last component matches a PascalCase enum",
			file: "order-status.proto",
		},
		{
			name: "lower_snake_case last component matches a PascalCase enum",
			file: "order-status-snake.proto",
		},
		{
			name: "last component that is not the enum name is flagged",
			file: "kind.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "Kind" must be declared in a package named after it, such as "app.test.kind", but the package is "app.test.enums".`,
					FileLocation: enumLocation("kind.proto", 0),
				},
			},
		},
		{
			name: "words joined without a separator do not match",
			file: "collapsed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "OrderStatus" must be declared in a package named after it, such as "app.test.orderStatus" or "app.test.order_status", but the package is "app.test.orderstatus".`,
					FileLocation: enumLocation("collapsed.proto", 0),
				},
			},
		},
		{
			name: "single-component package is suggested without a parent",
			file: "single.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "Kind" must be declared in a package named after it, such as "kind", but the package is "enums".`,
					FileLocation: enumLocation("single.proto", 0),
				},
			},
		},
		{
			name: "only the enum the package is not named after is flagged",
			file: "two.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "Status" must be declared in a package named after it, such as "app.test.status", but the package is "app.test.kind".`,
					FileLocation: enumLocation("two.proto", 1),
				},
			},
		},
		{
			name:    "only the camelCase form is suggested when PACKAGE_CAMEL_CASE is enabled",
			file:    "collapsed.proto",
			ruleIDs: []string{EnumDedicatedPackageRuleID, PackageCamelCaseRuleID},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "OrderStatus" must be declared in a package named after it, such as "app.test.orderStatus", but the package is "app.test.orderstatus".`,
					FileLocation: enumLocation("collapsed.proto", 0),
				},
			},
		},
		{
			name: "file with enums and messages is not checked",
			file: "mixed.proto",
		},
		{
			name: "file without a package is not checked",
			file: "no_package.proto",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ruleIDs := testCase.ruleIDs
			if ruleIDs == nil {
				ruleIDs = []string{EnumDedicatedPackageRuleID}
			}
			checktest.CheckTest{
				Request: &checktest.RequestSpec{
					Files: &checktest.ProtoFileSpec{
						DirPaths:  []string{enumDedicatedPackageTestdataDir},
						FilePaths: []string{testCase.file},
					},
					RuleIDs: ruleIDs,
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestEnumDedicatedPackageOtherFiles(t *testing.T) {
	t.Parallel()

	packageAnnotation := func(file, pkg, message string) checktest.ExpectedAnnotation {
		return checktest.ExpectedAnnotation{
			RuleID:  EnumDedicatedPackageRuleID,
			Message: message,
			FileLocation: &checktest.ExpectedFileLocation{
				FileName:  file,
				StartLine: 2,
				EndLine:   2,
				EndColumn: len("package " + pkg + ";"),
			},
		}
	}

	for _, testCase := range []struct {
		name                string
		files               []string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name:  "file with a message in the package of an enum file is flagged",
			files: []string{"color.proto", "palette.proto"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				packageAnnotation("palette.proto", "app.test.color", `Package "app.test.color" is dedicated to enum "Color" declared in "color.proto", so this file must be in another package.`),
			},
		},
		{
			name:  "enum file is found whatever the order of the files",
			files: []string{"palette.proto", "color.proto"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				packageAnnotation("palette.proto", "app.test.color", `Package "app.test.color" is dedicated to enum "Color" declared in "color.proto", so this file must be in another package.`),
			},
		},
		{
			name:  "file without declarations in the package of an enum file is flagged",
			files: []string{"color.proto", "options.proto"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				packageAnnotation("options.proto", "app.test.color", `Package "app.test.color" is dedicated to enum "Color" declared in "color.proto", so this file must be in another package.`),
			},
		},
		{
			name:  "enum file in the package of another enum file is flagged at the enum alone",
			files: []string{"color.proto", "shade.proto"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumDedicatedPackageRuleID,
					Message:      `Enum "Shade" must be declared in a package named after it, such as "app.test.shade", but the package is "app.test.color".`,
					FileLocation: enumLocation("shade.proto", 0),
				},
			},
		},
		{
			name:  "file in the parent package is accepted",
			files: []string{"color.proto", "mixed.proto"},
		},
		{
			name:  "package of a file with enums and messages is not dedicated",
			files: []string{"theme.proto", "theme-settings.proto"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			checktest.CheckTest{
				Request: &checktest.RequestSpec{
					Files: &checktest.ProtoFileSpec{
						DirPaths:  []string{enumDedicatedPackageTestdataDir},
						FilePaths: testCase.files,
					},
					RuleIDs: []string{EnumDedicatedPackageRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestSameWords(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		a, b     string
		expected bool
	}{
		{"Status", "status", true},
		{"OrderStatus", "orderStatus", true},
		{"OrderStatus", "order_status", true},
		{"OrderStatus", "ORDER_STATUS", true},
		{"HTTPStatus", "httpStatus", true},
		{"OrderStatus", "orderstatus", false},
		{"OrderStatus", "order", false},
		{"Status", "", false},
	} {
		t.Run(testCase.a+"/"+testCase.b, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, testCase.expected, sameWords(testCase.a, testCase.b))
		})
	}
}
