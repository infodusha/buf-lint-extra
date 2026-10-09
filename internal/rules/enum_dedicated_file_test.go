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
			name: "enums mixed with another enum, a message and a service are flagged once per enum",
			file: "mixed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Status" must be declared in a file of its own, but this file also declares 1 other enum, 1 message and 1 service.`,
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
					Message: `Enum "Role" must be declared in a file of its own, but this file also declares 1 other enum, 1 message and 1 service.`,
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
					Message: `Enum "Color" must be declared in a file of its own, but this file also declares 1 extension.`,
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
			name: "file with a single enum and file options is accepted",
			file: "enum_only.proto",
		},
		{
			name: "file with two enums is flagged once per enum",
			file: "two_enums.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Status" must be declared in a file of its own, but this file also declares 1 other enum.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "two_enums.proto",
						StartLine:   6,
						StartColumn: 0,
						EndLine:     8,
						EndColumn:   1,
					},
				},
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Role" must be declared in a file of its own, but this file also declares 1 other enum.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "two_enums.proto",
						StartLine:   10,
						StartColumn: 0,
						EndLine:     12,
						EndColumn:   1,
					},
				},
			},
		},
		{
			name: "enums nested in messages are flagged at any depth",
			file: "nested.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "User.Kind" must be declared at the top level of a file of its own, not nested in message "User".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "nested.proto",
						StartLine:   5,
						StartColumn: 2,
						EndLine:     7,
						EndColumn:   3,
					},
				},
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "User.Address.Type" must be declared at the top level of a file of its own, not nested in message "User.Address".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "nested.proto",
						StartLine:   10,
						StartColumn: 4,
						EndLine:     12,
						EndColumn:   5,
					},
				},
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Team.Visibility" must be declared at the top level of a file of its own, not nested in message "Team".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "nested.proto",
						StartLine:   23,
						StartColumn: 2,
						EndLine:     25,
						EndColumn:   3,
					},
				},
			},
		},
		{
			name: "top-level and nested enums in one file are both flagged",
			file: "mixed_nested.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "Status" must be declared in a file of its own, but this file also declares 1 message.`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "mixed_nested.proto",
						StartLine:   4,
						StartColumn: 0,
						EndLine:     6,
						EndColumn:   1,
					},
				},
				{
					RuleID:  EnumDedicatedFileRuleID,
					Message: `Enum "User.Kind" must be declared at the top level of a file of its own, not nested in message "User".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "mixed_nested.proto",
						StartLine:   9,
						StartColumn: 2,
						EndLine:     11,
						EndColumn:   3,
					},
				},
			},
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
