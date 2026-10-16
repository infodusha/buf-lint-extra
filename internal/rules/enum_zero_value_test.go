package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestEnumZeroValue(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_zero_value"

	valueAnnotation := func(file string, line, column int, value, message string) checktest.ExpectedAnnotation {
		return checktest.ExpectedAnnotation{
			RuleID:  EnumZeroValueRuleID,
			Message: message,
			FileLocation: &checktest.ExpectedFileLocation{
				FileName:    file,
				StartLine:   line,
				StartColumn: column,
				EndLine:     line,
				EndColumn:   column + len(value),
			},
		}
	}

	for _, testCase := range []struct {
		name                string
		file                string
		options             map[string]any
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "zero value with the enum name prefix is flagged",
			file: "status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("status.proto", 5, 2, "STATUS_UNSPECIFIED = 0;", `Enum zero value name "STATUS_UNSPECIFIED" should be "UNSPECIFIED".`),
			},
		},
		{
			name: "zero value named UNSPECIFIED is accepted",
			file: "color.proto",
		},
		{
			name: "nested enums at any depth are checked",
			file: "nested.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("nested.proto", 6, 4, "ROLE_UNSPECIFIED = 0;", `Enum zero value name "ROLE_UNSPECIFIED" should be "UNSPECIFIED".`),
				valueAnnotation("nested.proto", 11, 6, "THEME_UNSPECIFIED = 0;", `Enum zero value name "THEME_UNSPECIFIED" should be "UNSPECIFIED".`),
			},
		},
		{
			name: "every enum of a file is checked on its own",
			file: "mixed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("mixed.proto", 5, 2, "STATUS_UNSPECIFIED = 0;", `Enum zero value name "STATUS_UNSPECIFIED" should be "UNSPECIFIED".`),
			},
		},
		{
			name: "every alias of the zero value is checked",
			file: "alias.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("alias.proto", 8, 2, "STATUS_UNSPECIFIED = 0;", `Enum zero value name "STATUS_UNSPECIFIED" should be "UNSPECIFIED".`),
			},
		},
		{
			name: "non-zero values are not checked",
			file: "no_zero.proto",
		},
		{
			name: "file without a package is checked",
			file: "no_package.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("no_package.proto", 3, 2, "STATUS_UNSPECIFIED = 0;", `Enum zero value name "STATUS_UNSPECIFIED" should be "UNSPECIFIED".`),
			},
		},
		{
			name: "zero value other than the default name is flagged",
			file: "unknown.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("unknown.proto", 5, 2, "UNKNOWN = 0;", `Enum zero value name "UNKNOWN" should be "UNSPECIFIED".`),
			},
		},
		{
			name:    "zero value with the configured name is accepted",
			file:    "unknown.proto",
			options: map[string]any{EnumZeroValueOptionKey: "UNKNOWN"},
		},
		{
			name:    "zero value other than the configured name is flagged",
			file:    "color.proto",
			options: map[string]any{EnumZeroValueOptionKey: "UNKNOWN"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				valueAnnotation("color.proto", 5, 2, "UNSPECIFIED = 0;", `Enum zero value name "UNSPECIFIED" should be "UNKNOWN".`),
			},
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
					RuleIDs: []string{EnumZeroValueRuleID},
					Options: testCase.options,
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}
