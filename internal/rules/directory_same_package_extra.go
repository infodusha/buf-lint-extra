package rules

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"buf.build/go/bufplugin/check"
)

const DirectorySamePackageExtraRuleID = "DIRECTORY_SAME_PACKAGE_EXTRA"

var directorySamePackageExtraRule = newFilesRule(
	&check.RuleSpec{
		ID:      DirectorySamePackageExtraRuleID,
		Default: false,
		Purpose: "Checks that all files in a given directory are in the same package, like DIRECTORY_SAME_PACKAGE, counting enum files as files of their parent package when the component named after the enum is excluded from the directory.",
		Type:    check.RuleTypeLint,
	},
	checkDirectorySamePackageExtra,
)

func checkDirectorySamePackageExtra(files []fileSummary, request checkRequest) ([]fileAnnotation, error) {
	excludeEnumComponent, err := packageDirectoryExcludeEnumComponent(request)
	if err != nil {
		return nil, err
	}
	packagesByDir := map[string][]string{}
	for _, file := range files {
		dir := path.Dir(file.name)
		pkg := file.pkg
		if excludeEnumComponent {
			pkg = trimEnumComponent(pkg, file)
		}
		if !slices.Contains(packagesByDir[dir], pkg) {
			packagesByDir[dir] = append(packagesByDir[dir], pkg)
		}
	}
	var annotations []fileAnnotation
	for _, file := range files {
		dir := path.Dir(file.name)
		packages := packagesByDir[dir]
		if len(packages) < 2 {
			continue
		}
		var sourcePath []int32
		if file.pkg != "" {
			sourcePath = []int32{int32(fileDescriptorProtoPackage.number)}
		}
		annotations = append(annotations, fileAnnotation{
			fileName: file.name,
			annotation: annotation{
				message:    directorySamePackageMessage(packages, dir),
				sourcePath: sourcePath,
			},
		})
	}
	return annotations, nil
}

func directorySamePackageMessage(packages []string, dir string) string {
	packages = slices.Sorted(slices.Values(packages))
	noPackage := packages[0] == ""
	if noPackage {
		packages = packages[1:]
	}
	message := fmt.Sprintf("Package %q", packages[0])
	if len(packages) > 1 {
		message = fmt.Sprintf("Multiple packages %q", strings.Join(packages, ","))
	}
	if noPackage {
		message += " and file with no package"
	}
	return fmt.Sprintf("%s detected within directory %q.", message, dir)
}
