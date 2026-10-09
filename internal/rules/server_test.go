package rules

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"buf.build/go/bufplugin/info"
	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"pluginrpc.com/pluginrpc"
)

type annotationSummary struct {
	RuleID     string
	Message    string
	FileName   string
	SourcePath protoreflect.SourcePath
	StartLine  int
	EndLine    int
}

func TestServerMatchesSpec(t *testing.T) {
	t.Parallel()

	specClient, err := check.NewClientForSpec(Spec)
	require.NoError(t, err)
	serverClient := check.NewClient(newTestPluginrpcClient(t))

	var ruleIDs []string
	for _, rule := range allRules {
		ruleIDs = append(ruleIDs, rule.spec.ID)
	}
	dirs, err := filepath.Glob("testdata/*")
	require.NoError(t, err)
	for _, dir := range dirs {
		var files []string
		require.NoError(t, fs.WalkDir(os.DirFS(dir), ".", func(path string, entry fs.DirEntry, err error) error {
			if err == nil && !entry.IsDir() && filepath.Ext(path) == ".proto" {
				files = append(files, path)
			}
			return err
		}))
		fileSets := make([][]string, 0, len(files)+1)
		for _, file := range files {
			fileSets = append(fileSets, []string{file})
		}
		if dir == directorySamePackageExtraTestdataDir {
			fileSets = append(fileSets, files)
		}
		for _, fileSet := range fileSets {
			for _, variant := range []struct {
				name          string
				ruleIDs       []string
				options       map[string]any
				errorContains string
			}{
				{name: "default rules"},
				{name: "all rules", ruleIDs: ruleIDs},
				{
					name:    "all rules with options",
					ruleIDs: ruleIDs,
					options: map[string]any{
						EnumFileSuffixOptionKey:                   "_enums",
						PackageDirectoryExcludedPrefixesOptionKey: []string{"test", "legacy"},
						PackageDirectoryCaseOptionKey:             "lower-kebab-case",
						PackageDirectoryEnumComponentOptionKey:    "excluded",
					},
				},
				{
					name:          "invalid suffix option",
					ruleIDs:       ruleIDs,
					options:       map[string]any{EnumFileSuffixOptionKey: ".proto"},
					errorContains: EnumFileSuffixOptionKey,
				},
				{
					name:          "invalid case option",
					ruleIDs:       ruleIDs,
					options:       map[string]any{PackageDirectoryCaseOptionKey: "UPPER"},
					errorContains: PackageDirectoryCaseOptionKey,
				},
			} {
				t.Run(filepath.Join(dir, strings.Join(fileSet, "+"))+"/"+variant.name, func(t *testing.T) {
					t.Parallel()
					ctx := t.Context()
					request, err := (&checktest.RequestSpec{
						Files: &checktest.ProtoFileSpec{
							DirPaths:  []string{dir},
							FilePaths: fileSet,
						},
						RuleIDs: variant.ruleIDs,
						Options: variant.options,
					}).ToRequest(ctx)
					require.NoError(t, err)
					specResponse, specErr := specClient.Check(ctx, request)
					serverResponse, serverErr := serverClient.Check(ctx, request)
					if variant.errorContains != "" {
						require.ErrorContains(t, specErr, variant.errorContains)
						require.ErrorContains(t, serverErr, variant.errorContains)
						return
					}
					require.NoError(t, specErr)
					require.NoError(t, serverErr)
					require.Equal(t, summarizeAnnotations(specResponse), summarizeAnnotations(serverResponse))
				})
			}
		}
	}
}

func TestServerProcedures(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	specServer, err := check.NewServer(Spec)
	require.NoError(t, err)
	specPluginrpcClient := pluginrpc.NewClient(pluginrpc.NewServerRunner(specServer))
	serverPluginrpcClient := newTestPluginrpcClient(t)

	specSpec, err := specPluginrpcClient.Spec(ctx)
	require.NoError(t, err)
	serverSpec, err := serverPluginrpcClient.Spec(ctx)
	require.NoError(t, err)
	require.True(t, proto.Equal(pluginrpc.NewProtoSpec(specSpec), pluginrpc.NewProtoSpec(serverSpec)))

	specRules, err := check.NewClient(specPluginrpcClient).ListRules(ctx)
	require.NoError(t, err)
	serverRules, err := check.NewClient(serverPluginrpcClient).ListRules(ctx)
	require.NoError(t, err)
	require.Equal(t, ruleIDsForRules(specRules), ruleIDsForRules(serverRules))

	specCategories, err := check.NewClient(specPluginrpcClient).ListCategories(ctx)
	require.NoError(t, err)
	serverCategories, err := check.NewClient(serverPluginrpcClient).ListCategories(ctx)
	require.NoError(t, err)
	require.Equal(t, categoryIDsForCategories(specCategories), categoryIDsForCategories(serverCategories))
	require.Equal(t, []string{ExtraCategoryID}, categoryIDsForCategories(serverCategories))
	for _, rule := range serverRules {
		require.Equal(t, []string{ExtraCategoryID}, categoryIDsForCategories(rule.Categories()))
	}

	pluginInfo, err := info.NewClient(serverPluginrpcClient).GetPluginInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, Spec.Info.Documentation, pluginInfo.Documentation())
}

func newTestPluginrpcClient(t *testing.T) pluginrpc.Client {
	t.Helper()
	server, err := newServer(protovalidate.GlobalValidator)
	require.NoError(t, err)
	return pluginrpc.NewClient(pluginrpc.NewServerRunner(server))
}

func summarizeAnnotations(response check.Response) []annotationSummary {
	var summaries []annotationSummary
	for _, annotation := range response.Annotations() {
		location := annotation.FileLocation()
		summaries = append(summaries, annotationSummary{
			RuleID:     annotation.RuleID(),
			Message:    annotation.Message(),
			FileName:   location.FileDescriptor().ProtoreflectFileDescriptor().Path(),
			SourcePath: location.SourcePath(),
			StartLine:  location.StartLine(),
			EndLine:    location.EndLine(),
		})
	}
	return summaries
}

func categoryIDsForCategories(categories []check.Category) []string {
	ids := make([]string, len(categories))
	for i, category := range categories {
		ids[i] = category.ID()
	}
	return ids
}

func ruleIDsForRules(rules []check.Rule) []string {
	ids := make([]string, len(rules))
	for i, rule := range rules {
		ids[i] = rule.ID()
	}
	return ids
}
