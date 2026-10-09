package rules

import (
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestPackageDirectoryMatchExtra(t *testing.T) {
	t.Parallel()

	const testdataDir = "testdata/package_directory_match_extra"

	for _, testCase := range []struct {
		name                string
		file                string
		options             map[string]any
		expectedAnnotations []checktest.ExpectedAnnotation
	}{
		{
			name: "file in the matching directory is accepted",
			file: "acme/v1/matching.proto",
		},
		{
			name: "file in another directory is flagged",
			file: "acme/misplaced.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "acme.v1" must be within a directory "acme/v1" relative to root but were in directory "acme".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "acme/misplaced.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   16,
					},
				},
			},
		},
		{
			name: "file in the root directory is flagged",
			file: "root.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "acme.v1" must be within a directory "acme/v1" relative to root but were in directory ".".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "root.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   16,
					},
				},
			},
		},
		{
			name: "file without a package is accepted",
			file: "no_package.proto",
		},
		{
			name: "prefix is required in the directory by default",
			file: "billing/v1/invoice.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "legacy.billing.v1" must be within a directory "legacy/billing/v1" relative to root but were in directory "billing/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "billing/v1/invoice.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   26,
					},
				},
			},
		},
		{
			name:    "excluded prefix is not required in the directory",
			file:    "billing/v1/invoice.proto",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy"}},
		},
		{
			name:    "prefix matches whole components only",
			file:    "billing/v1/invoice.proto",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy.bill"}},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "legacy.billing.v1" must be within a directory "legacy/billing/v1" relative to root but were in directory "billing/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "billing/v1/invoice.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   26,
					},
				},
			},
		},
		{
			name:    "longest excluded prefix wins",
			file:    "billing/v1/invoice.proto",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy", "legacy.billing"}},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "legacy.billing.v1" must be within a directory "v1" relative to root but were in directory "billing/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "billing/v1/invoice.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   26,
					},
				},
			},
		},
		{
			name:    "package that is an excluded prefix is not checked",
			file:    "anywhere/legacy.proto",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy"}},
		},
		{
			name: "camelCase component requires a camelCase directory by default",
			file: "acme/main-goal/v1/goal.proto",
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "acme.mainGoal.v1" must be within a directory "acme/mainGoal/v1" relative to root but were in directory "acme/main-goal/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "acme/main-goal/v1/goal.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   25,
					},
				},
			},
		},
		{
			name:    "lower-kebab-case directory matches a camelCase component",
			file:    "acme/main-goal/v1/goal.proto",
			options: map[string]any{PackageDirectoryCaseOptionKey: "lower-kebab-case"},
		},
		{
			name:    "camelCase directory does not match with lower-kebab-case conversion",
			file:    "acme/mainGoal/v1/goal.proto",
			options: map[string]any{PackageDirectoryCaseOptionKey: "lower-kebab-case"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "acme.mainGoal.v1" must be within a directory "acme/main-goal/v1" relative to root but were in directory "acme/mainGoal/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "acme/mainGoal/v1/goal.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   25,
					},
				},
			},
		},
		{
			name:    "lower_snake_case conversion expects underscores",
			file:    "acme/main-goal/v1/goal.proto",
			options: map[string]any{PackageDirectoryCaseOptionKey: "lower_snake_case"},
			expectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  PackageDirectoryMatchExtraRuleID,
					Message: `Files with package "acme.mainGoal.v1" must be within a directory "acme/main_goal/v1" relative to root but were in directory "acme/main-goal/v1".`,
					FileLocation: &checktest.ExpectedFileLocation{
						FileName:    "acme/main-goal/v1/goal.proto",
						StartLine:   2,
						StartColumn: 0,
						EndLine:     2,
						EndColumn:   25,
					},
				},
			},
		},
		{
			name: "excluded prefix and case conversion combine",
			file: "billing/order-items/v1/item.proto",
			options: map[string]any{
				PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy"},
				PackageDirectoryCaseOptionKey:             "lower-kebab-case",
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
					RuleIDs: []string{PackageDirectoryMatchExtraRuleID},
					Options: testCase.options,
				},
				Spec:                Spec,
				ExpectedAnnotations: testCase.expectedAnnotations,
			}.Run(t)
		})
	}
}

func TestPackageDirectoryMatchExtraInvalidOptions(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name    string
		options map[string]any
		key     string
	}{
		{
			name:    "unknown case",
			options: map[string]any{PackageDirectoryCaseOptionKey: "UPPER"},
			key:     PackageDirectoryCaseOptionKey,
		},
		{
			name:    "non-string case",
			options: map[string]any{PackageDirectoryCaseOptionKey: true},
			key:     PackageDirectoryCaseOptionKey,
		},
		{
			name:    "prefixes as a single string",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: "legacy"},
			key:     PackageDirectoryExcludedPrefixesOptionKey,
		},
		{
			name:    "prefix with a trailing dot",
			options: map[string]any{PackageDirectoryExcludedPrefixesOptionKey: []string{"legacy."}},
			key:     PackageDirectoryExcludedPrefixesOptionKey,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			request, err := (&checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths:  []string{"testdata/package_directory_match_extra"},
					FilePaths: []string{"no_package.proto"},
				},
				RuleIDs: []string{PackageDirectoryMatchExtraRuleID},
				Options: testCase.options,
			}).ToRequest(ctx)
			require.NoError(t, err)
			client, err := check.NewClientForSpec(Spec)
			require.NoError(t, err)
			_, err = client.Check(ctx, request)
			require.ErrorContains(t, err, testCase.key)
		})
	}
}

func TestTrimPackagePrefix(t *testing.T) {
	t.Parallel()

	prefixes := []string{"acme", "acme.v1", "legacy"}
	for input, expected := range map[string]string{
		"":                "",
		"acme":            "",
		"acme.v1":         "",
		"acme.v1.billing": "billing",
		"acme.v2.billing": "v2.billing",
		"acmeco.v1":       "acmeco.v1",
		"legacy.acme.v1":  "acme.v1",
		"other.acme.v1":   "other.acme.v1",
		"billing.acme.v1": "billing.acme.v1",
	} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, expected, trimPackagePrefix(input, prefixes))
		})
	}
}
