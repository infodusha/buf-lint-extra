package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestEnumValueNoPrefix(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_value_no_prefix"

	valueAnnotation := func(file string, line int, value, message string) checktest.ExpectedAnnotation {
		return checktest.ExpectedAnnotation{
			RuleID:  EnumValueNoPrefixRuleID,
			Message: message,
			FileLocation: &checktest.ExpectedFileLocation{
				FileName:    file,
				StartLine:   line,
				StartColumn: 2,
				EndLine:     line,
				EndColumn:   2 + len(value),
			},
		}
	}

	for _, testCase := range []struct {
		name                string
		file                string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "values prefixed with the enum name are flagged",
			file: "status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("status.proto", 5, "STATUS_UNSPECIFIED = 0;", `Enum value name "STATUS_UNSPECIFIED" should not be prefixed with "STATUS_", such as "UNSPECIFIED".`),
				valueAnnotation("status.proto", 6, "STATUS_ACTIVE = 1;", `Enum value name "STATUS_ACTIVE" should not be prefixed with "STATUS_", such as "ACTIVE".`),
			},
		},
		{
			name: "prefix is the UPPER_SNAKE_CASE form of the enum name",
			file: "order-status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("order-status.proto", 5, "ORDER_STATUS_UNSPECIFIED = 0;", `Enum value name "ORDER_STATUS_UNSPECIFIED" should not be prefixed with "ORDER_STATUS_", such as "UNSPECIFIED".`),
			},
		},
		{
			name: "acronyms in the enum name are split into words",
			file: "http-status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("http-status.proto", 5, "HTTP_STATUS_UNSPECIFIED = 0;", `Enum value name "HTTP_STATUS_UNSPECIFIED" should not be prefixed with "HTTP_STATUS_", such as "UNSPECIFIED".`),
			},
		},
		{
			name: "value that is only the prefix is flagged without a suggestion",
			file: "bare.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("bare.proto", 5, "STATUS_ = 0;", `Enum value name "STATUS_" should not be prefixed with "STATUS_".`),
			},
		},
		{
			name: "only the enum the package is named after is checked",
			file: "two.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("two.proto", 5, "STATUS_UNSPECIFIED = 0;", `Enum value name "STATUS_UNSPECIFIED" should not be prefixed with "STATUS_", such as "UNSPECIFIED".`),
			},
		},
		{
			name: "values without the prefix are accepted",
			file: "color.proto",
		},
		{
			name: "enum in a package not named after it is not checked",
			file: "kind.proto",
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
			checktest.CheckTest{
				Request: &checktest.RequestSpec{
					Files: &checktest.ProtoFileSpec{
						DirPaths:  []string{testdataDir},
						FilePaths: []string{testCase.file},
					},
					RuleIDs: []string{EnumValueNoPrefixRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}
