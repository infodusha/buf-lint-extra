package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestFileLowerKebabCase(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/file_lower_kebab_case"

	for _, testCase := range []struct {
		name                string
		file                string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "lower_snake_case file name is flagged",
			file: "user_service.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       FileLowerKebabCaseRuleID,
					Message:      `Filename "user_service.proto" should be lower-kebab-case.proto, such as "user-service.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "user_service.proto"},
				},
			},
		},
		{
			name: "PascalCase file name is flagged",
			file: "UserService.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       FileLowerKebabCaseRuleID,
					Message:      `Filename "UserService.proto" should be lower-kebab-case.proto, such as "user-service.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "UserService.proto"},
				},
			},
		},
		{
			name: "file in a nested directory is reported by its base name",
			file: "acme/v1/user_profile.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       FileLowerKebabCaseRuleID,
					Message:      `Filename "user_profile.proto" should be lower-kebab-case.proto, such as "user-profile.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "acme/v1/user_profile.proto"},
				},
			},
		},
		{
			name: "segment before a dotted suffix is flagged on its own",
			file: "User_Status.enum.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       FileLowerKebabCaseRuleID,
					Message:      `Filename "User_Status.enum.proto" should be lower-kebab-case.proto, such as "user-status.enum.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "User_Status.enum.proto"},
				},
			},
		},
		{
			name: "empty segments are dropped from the suggestion",
			file: "user..status.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:       FileLowerKebabCaseRuleID,
					Message:      `Filename "user..status.proto" should be lower-kebab-case.proto, such as "user.status.proto".`,
					FileLocation: &checktest.ExpectedFileLocation{FileName: "user..status.proto"},
				},
			},
		},
		{
			name: "lower-kebab-case file name is accepted",
			file: "user-service.proto",
		},
		{
			name: "dotted suffix is a segment of its own",
			file: "user-status.enum.proto",
		},
		{
			name: "single-word file name is accepted",
			file: "user.proto",
		},
		{
			name: "directories are not checked",
			file: "user_service/v1/user-service.proto",
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
					RuleIDs: []string{FileLowerKebabCaseRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestToLowerKebabCase(t *testing.T) {
	t.Parallel()

	for input, expected := range map[string]string{
		"":               "",
		"user":           "user",
		"v1beta1":        "v1beta1",
		"oauth2client":   "oauth2client",
		"user-service":   "user-service",
		"user_service":   "user-service",
		"user--service":  "user-service",
		"-user-service-": "user-service",
		"userService":    "user-service",
		"UserService":    "user-service",
		"USER_SERVICE":   "user-service",
		"userV2":         "user-v2",
		"userAPI":        "user-api",
		"HTTPServer":     "http-server",
	} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, expected, toLowerKebabCase(input))
		})
	}
}
