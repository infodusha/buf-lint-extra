package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestEnumFileSuffixDefault(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_file_suffix"

	for _, testCase := range []struct {
		name                string
		file                string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "enum file without the suffix is flagged",
			file: "status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileSuffixRuleID,
					Message:      `File "status.proto" declares top-level enums and must have a name ending in "_enum.proto", such as "status_enum.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "status.proto"},
				},
			},
		},
		{
			name: "enum file in a nested directory keeps its directory in the suggestion",
			file: "acme/v1/kind.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileSuffixRuleID,
					Message:      `File "acme/v1/kind.proto" declares top-level enums and must have a name ending in "_enum.proto", such as "acme/v1/kind_enum.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "acme/v1/kind.proto"},
				},
			},
		},
		{
			name: "enum file with other declarations is told to move the enums instead of renaming",
			file: "mixed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileSuffixRuleID,
					Message:      `File "mixed.proto" declares top-level enums alongside 1 message and 1 service, so the enums must move to a file with a name ending in "_enum.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "mixed.proto"},
				},
			},
		},
		{
			name: "file with the suffix but without enums is flagged",
			file: "color_enum.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       EnumFileSuffixRuleID,
					Message:      `File "color_enum.proto" has a name ending in "_enum.proto" but declares no top-level enums.`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "color_enum.proto"},
				},
			},
		},
		{
			name: "enum file with the suffix is accepted",
			file: "role_enum.proto",
		},
		{
			name: "message file without the suffix is accepted",
			file: "user.proto",
		},
		{
			name: "nested enums do not require the suffix",
			file: "wrapper.proto",
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
					RuleIDs: []string{EnumFileSuffixRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestEnumFileSuffixOption(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/enum_file_suffix_option"

	// The option is accepted both with and without the ".proto" extension.
	for _, optionValue := range []string{"_enums", "_enums.proto"} {
		for _, testCase := range []struct {
			name                string
			file                string
			expectedAnnotations []checktest.ExpectedAnnotation
		}{
			{
				name: "enum file without the configured suffix is flagged",
				file: "status.proto",
				expectedAnnotations: []checktest.ExpectedAnnotation{
					{
						RuleID:       EnumFileSuffixRuleID,
						Message:      `File "status.proto" declares top-level enums and must have a name ending in "_enums.proto", such as "status_enums.proto".`,
						FileLocation: &checktest.ExpectedFileLocation{FileName: "status.proto"},
					},
				},
			},
			{
				name: "file with the configured suffix but without enums is flagged",
				file: "color_enums.proto",
				expectedAnnotations: []checktest.ExpectedAnnotation{
					{
						RuleID:       EnumFileSuffixRuleID,
						Message:      `File "color_enums.proto" has a name ending in "_enums.proto" but declares no top-level enums.`,
						FileLocation: &checktest.ExpectedFileLocation{FileName: "color_enums.proto"},
					},
				},
			},
			{
				name: "enum file with the configured suffix is accepted",
				file: "kind_enums.proto",
			},
		} {
			t.Run(optionValue+"/"+testCase.name, func(t *testing.T) {
				t.Parallel()
				checktest.CheckTest{
					Request: &checktest.RequestSpec{
						Files: &checktest.ProtoFileSpec{
							DirPaths:  []string{testdataDir},
							FilePaths: []string{testCase.file},
						},
						RuleIDs: []string{EnumFileSuffixRuleID},
						Options: map[string]any{EnumFileSuffixOptionKey: optionValue},
					},
					Spec:                Spec,
					ExpectedAnnotations: testCase.expectedAnnotations,
				}.Run(t)
			})
		}
	}
}

func TestEnumFileSuffixWithEnumDedicatedFile(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/enum_file_suffix"},
				FilePaths: []string{"mixed.proto"},
			},
			RuleIDs: []string{EnumDedicatedFileRuleID, EnumFileSuffixRuleID},
		},
		Spec: Spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  EnumDedicatedFileRuleID,
				Message: `Enum "Status" must be declared in a dedicated file that contains only enums, but this file also declares 1 message and 1 service.`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "mixed.proto",
					StartLine:   4,
					StartColumn: 0,
					EndLine:     6,
					EndColumn:   1,
				},
			},
			{
				RuleID:       EnumFileSuffixRuleID,
				Message:      `File "mixed.proto" declares top-level enums alongside 1 message and 1 service, so the enums must move to a file with a name ending in "_enum.proto".`,
				FileLocation: &checktest.ExpectedFileLocation{FileName: "mixed.proto"},
			},
		},
	}.Run(t)
}

func TestEnumFileSuffixInvalidOption(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		optionValue any
	}{
		{name: "non-string value", optionValue: int64(1)},
		{name: "only the extension", optionValue: ".proto"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			request, err := (&checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths:  []string{"testdata/enum_file_suffix"},
					FilePaths: []string{"user.proto"},
				},
				RuleIDs: []string{EnumFileSuffixRuleID},
				Options: map[string]any{EnumFileSuffixOptionKey: testCase.optionValue},
			}).ToRequest(ctx)
			require.NoError(t, err)
			client, err := check.NewClientForSpec(Spec)
			require.NoError(t, err)
			_, err = client.Check(ctx, request)
			require.Error(t, err)
			require.ErrorContains(t, err, EnumFileSuffixOptionKey)
		})
	}
}
