package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestEnumDedicatedFile(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_dedicated_file"

	for _, testCase := range []struct {
		name                string
		file                string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "enums mixed with a message and a service are flagged once per enum",
			file: "mixed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Status" must be declared in a dedicated file that contains only enums, but this file also declares 1 message and 1 service.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "mixed.proto",
						StartLine:   4,
						StartColumn: 0,
						EndLine:     7,
						EndColumn:   1,
					},
				},
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Role" must be declared in a dedicated file that contains only enums, but this file also declares 1 message and 1 service.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "mixed.proto",
						StartLine:   13,
						StartColumn: 0,
						EndLine:     15,
						EndColumn:   1,
					},
				},
			},
		},
		{
			name: "enum mixed with an extension is flagged",
			file: "with_extension.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Color" must be declared in a dedicated file that contains only enums, but this file also declares 1 extension.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "with_extension.proto",
						StartLine:   6,
						StartColumn: 0,
						EndLine:     8,
						EndColumn:   1,
					},
				},
			},
		},
		{
			name: "file with only enums and file options is accepted",
			file: "enum_only.proto",
		},
		{
			name: "nested enums do not make a message file an enum file",
			file: "message_only.proto",
		},
		{
			name: "mixed declarations in imported files are not reported",
			file: "imports_mixed.proto",
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
					RuleIDs: []string{EnumDedicatedFileRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}
