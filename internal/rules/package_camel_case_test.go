package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestPackageCamelCase(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/package_camel_case"

	for _, testCase := range []struct {
		name                string
		file                string
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "lower_snake_case component is flagged",
			file: "snake.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageCamelCaseRuleID,
					Message: `Package name "acme.user_service.v1" should be camelCase, such as "acme.userService.v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "snake.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   29,
					},
				},
			},
		},
		{
			name: "PascalCase component is flagged",
			file: "pascal.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageCamelCaseRuleID,
					Message: `Package name "acme.UserService.v1" should be camelCase, such as "acme.userService.v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "pascal.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   28,
					},
				},
			},
		},
		{
			name: "every component is converted in the suggestion",
			file: "mixed.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageCamelCaseRuleID,
					Message: `Package name "Acme.HTTP_gateway.V1" should be camelCase, such as "acme.httpGateway.v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "mixed.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   29,
					},
				},
			},
		},
		{
			name: "camelCase package with a version suffix is accepted",
			file: "camel.proto",
		},
		{
			name: "uppercase acronym after the first word is accepted",
			file: "acronym.proto",
		},
		{
			name: "single-word components are accepted",
			file: "single.proto",
		},
		{
			name: "file without a package is accepted",
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
					RuleIDs: []string{PackageCamelCaseRuleID},
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	t.Parallel()

	for input, expected := range map[string]string{
		"":                "",
		"v1":              "v1",
		"v1beta1":         "v1beta1",
		"V1":              "v1",
		"user":            "user",
		"User":            "user",
		"USER":            "user",
		"user_service":    "userService",
		"userService":     "userService",
		"UserService":     "userService",
		"user__service":   "userService",
		"_user_service_":  "userService",
		"user_service_v2": "userServiceV2",
		"userAPI":         "userAPI",
		"HTTPServer":      "httpServer",
		"HTTP_server":     "httpServer",
		"http_SERVER":     "httpSERVER",
		"oauth2_client":   "oauth2Client",
	} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, expected, toCamelCase(input))
		})
	}
}
