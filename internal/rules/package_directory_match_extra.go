package rules

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const (
	PackageDirectoryMatchExtraRuleID = "PACKAGE_DIRECTORY_MATCH_EXTRA"

	PackageDirectoryExcludedPrefixesOptionKey = "package_directory_excluded_prefixes"
	PackageDirectoryCaseOptionKey             = "package_directory_case"
	PackageDirectoryEnumComponentOptionKey    = "package_directory_enum_component"
)

var packageDirectoryCases = map[string]func(string) string{
	"lower-kebab-case": toLowerKebabCase,
	"lower_snake_case": toLowerSnakeCase,
}

var packageDirectoryMatchExtraRule = newRule(
	&check.RuleSpec{
		ID:      PackageDirectoryMatchExtraRuleID,
		Default: false,
		Purpose: "Checks that files are in a directory matching their package, like PACKAGE_DIRECTORY_MATCH, with options to exclude package prefixes, to convert the components to a case, and to exclude the component named after the enum of an enum file.",
		Type:    check.RuleTypeLint,
	},
	checkPackageDirectoryMatchExtra,
)

func checkPackageDirectoryMatchExtra(file fileSummary, request checkRequest) ([]annotation, error) {
	prefixes, err := packageDirectoryExcludedPrefixes(request.options)
	if err != nil {
		return nil, err
	}
	convert, err := packageDirectoryCase(request.options)
	if err != nil {
		return nil, err
	}
	excludeEnumComponent, err := packageDirectoryExcludeEnumComponent(request)
	if err != nil {
		return nil, err
	}
	pkg := trimPackagePrefix(file.pkg, prefixes)
	if excludeEnumComponent {
		pkg = trimEnumComponent(pkg, file)
	}
	if pkg == "" {
		return nil, nil
	}
	components := strings.Split(pkg, ".")
	for i, component := range components {
		components[i] = convert(component)
	}
	expectedDir := strings.Join(components, "/")
	dir := path.Dir(file.name)
	if dir == expectedDir {
		return nil, nil
	}
	return []annotation{{
		message: fmt.Sprintf(
			"Files with package %q must be within a directory %q relative to root but were in directory %q.",
			file.pkg,
			expectedDir,
			dir,
		),
		sourcePath: []int32{int32(fileDescriptorProtoPackage.number)},
	}}, nil
}

func packageDirectoryExcludedPrefixes(options option.Options) ([]string, error) {
	prefixes, err := option.GetStringSliceValue(options, PackageDirectoryExcludedPrefixesOptionKey)
	if err != nil {
		return nil, err
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(prefix, ".") || strings.HasSuffix(prefix, ".") {
			return nil, fmt.Errorf("option %q must contain package prefixes such as \"acme.v1\", got %q", PackageDirectoryExcludedPrefixesOptionKey, prefix)
		}
	}
	return prefixes, nil
}

// packageDirectoryExcludeEnumComponent reports whether the component named
// after the enum is left out of the directory. Unless the option says so, it
// is left out exactly when ENUM_DEDICATED_PACKAGE runs in the same request,
// since that rule gives every enum file such a component.
func packageDirectoryExcludeEnumComponent(request checkRequest) (bool, error) {
	value, err := option.GetStringValue(request.options, PackageDirectoryEnumComponentOptionKey)
	if err != nil {
		return false, err
	}
	switch value {
	case "":
		return request.enables(enumDedicatedPackageRule), nil
	case "included":
		return false, nil
	case "excluded":
		return true, nil
	default:
		return false, fmt.Errorf("option %q must be \"included\" or \"excluded\", got %q", PackageDirectoryEnumComponentOptionKey, value)
	}
}

func packageDirectoryCase(options option.Options) (func(string) string, error) {
	value, err := option.GetStringValue(options, PackageDirectoryCaseOptionKey)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return func(component string) string { return component }, nil
	}
	convert, ok := packageDirectoryCases[value]
	if !ok {
		names := slices.Sorted(maps.Keys(packageDirectoryCases))
		for i, name := range names {
			names[i] = strconv.Quote(name)
		}
		return nil, fmt.Errorf("option %q must be %s, got %q", PackageDirectoryCaseOptionKey, strings.Join(names, " or "), value)
	}
	return convert, nil
}

// trimPackagePrefix removes the longest of the prefixes that matches whole
// components of pkg, so "acme.v1.billing" with the prefix "acme.v1" becomes
// "billing", and "" when the whole package is a prefix.
func trimPackagePrefix(pkg string, prefixes []string) string {
	longest := ""
	for _, prefix := range prefixes {
		if len(prefix) > len(longest) && (pkg == prefix || strings.HasPrefix(pkg, prefix+".")) {
			longest = prefix
		}
	}
	return strings.TrimPrefix(strings.TrimPrefix(pkg, longest), ".")
}

// trimEnumComponent removes the last component of pkg when file declares
// only enums and one of them is named after that component, so the package
// ENUM_DEDICATED_PACKAGE asks for becomes its parent.
func trimEnumComponent(pkg string, file fileSummary) string {
	if parent, last := splitLastComponent(pkg); file.declaresOnlyEnums() && file.declaresEnumNamed(last) {
		return parent
	}
	return pkg
}
