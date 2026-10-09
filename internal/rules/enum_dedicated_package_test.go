package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestEnumDedicatedPackage(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_dedicated_package"

	enumLocation := func(file string, index int) *checktest.ExpectedFileLocation {
		return &checktest.ExpectedFileLocation{
			FileName:    file,
			StartLine:   4 + 4*index,
			StartColumn: 0,
			EndLine:     6 + 4*index,
			EndColumn:   1,
		}
	}

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
					Message:      `Enum "Kind" must be declared in a package named after it, such as "app.test.kind", but the package is "app.test.status".`,
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
						DirPaths:  []string{testdataDir},
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
