package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestEnumFileMatch(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_file_match"

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
		options             map[string]any
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "file named after the enum is accepted",
			file: "status.proto",
		},
		{
			name: "lower-kebab-case file name matches a PascalCase enum",
			file: "order-status.proto",
		},
		{
			name: "lower_snake_case file name matches a PascalCase enum",
			file: "order_status.proto",
		},
		{
			name: "default enum file suffix is allowed",
			file: "status_enum.proto",
		},
		{
			name: "enum whose name ends in the suffix matches the whole file name",
			file: "status-enum.proto",
		},
		{
			name:    "configured enum file suffix is allowed",
			file:    "user-status.enum.proto",
			options: map[string]any{EnumFileSuffixOptionKey: ".enum"},
		},
		{
			name: "suffix that is not configured is part of the name",
			file: "user-status.enum.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "UserStatus" must be declared in a file named after it, such as "user-status.proto" or "user_status.proto", but the file is "user-status.enum.proto".`,
					FileLocation: enumLocation("user-status.enum.proto", 0),
				},
			},
		},
		{
			name: "file not named after the enum is flagged",
			file: "enums.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "Kind" must be declared in a file named after it, such as "kind.proto", but the file is "enums.proto".`,
					FileLocation: enumLocation("enums.proto", 0),
				},
			},
		},
		{
			name: "suffix of the file is kept in the suggestion",
			file: "enums_enum.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "Kind" must be declared in a file named after it, such as "kind_enum.proto", but the file is "enums_enum.proto".`,
					FileLocation: enumLocation("enums_enum.proto", 0),
				},
			},
		},
		{
			name: "words joined without a separator do not match",
			file: "orderstatus.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "OrderStatus" must be declared in a file named after it, such as "order-status.proto" or "order_status.proto", but the file is "orderstatus.proto".`,
					FileLocation: enumLocation("orderstatus.proto", 0),
				},
			},
		},
		{
			name: "directory is kept in the suggestion",
			file: "acme/v1/enums.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "Kind" must be declared in a file named after it, such as "acme/v1/kind.proto", but the file is "acme/v1/enums.proto".`,
					FileLocation: enumLocation("acme/v1/enums.proto", 0),
				},
			},
		},
		{
			name: "only the enum the file is not named after is flagged",
			file: "kind.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileMatchRuleID,
					Message:      `Enum "Status" must be declared in a file named after it, such as "status.proto", but the file is "kind.proto".`,
					FileLocation: enumLocation("kind.proto", 0),
				},
			},
		},
		{
			name: "file with enums and messages is not checked",
			file: "mixed.proto",
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
					RuleIDs: []string{EnumFileMatchRuleID},
					Options: testCase.options,
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestEnumFileMatchInvalidSuffixOption(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	request, err := (&checktest.RequestSpec{
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{"testdata/enum_file_match"},
			FilePaths: []string{"status.proto"},
		},
		RuleIDs: []string{EnumFileMatchRuleID},
		Options: map[string]any{EnumFileSuffixOptionKey: ".proto"},
	}).ToRequest(ctx)
	require.NoError(t, err)
	client, err := check.NewClientForSpec(Spec)
	require.NoError(t, err)
	_, err = client.Check(ctx, request)
	require.ErrorContains(t, err, EnumFileSuffixOptionKey)
}
