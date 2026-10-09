package rules

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"buf.build/go/bufplugin/check"
)

const EnumDedicatedPackageRuleID = "ENUM_DEDICATED_PACKAGE"

var enumDedicatedPackageRule = newRule(
	&check.RuleSpec{
		ID:      EnumDedicatedPackageRuleID,
		Default: false,
		Purpose: "Checks that files declaring only enums have a package whose last component is the name of the enum, in any case style, so that every enum gets a package of its own.",
		Type:    check.RuleTypeLint,
	},
	checkEnumDedicatedPackage,
)

func checkEnumDedicatedPackage(file fileSummary, _ checkRequest) ([]annotation, error) {
	if file.pkg == "" || !file.declaresOnlyEnums() {
		return nil, nil
	}
	parent, last := splitLastComponent(file.pkg)
	var annotations []annotation
	for i, enum := range file.enums {
		if sameWords(enum, last) {
			continue
		}
		annotations = append(annotations, annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared in a package named after it, such as %s, but the package is %q.",
				enum,
				packagesNamedAfter(parent, enum),
				file.pkg,
			),
			sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
		})
	}
	return annotations, nil
}

// packagesNamedAfter lists the packages under parent whose last component is
// enum in camelCase and in lower_snake_case, quoted and joined with "or", or
// a single package when both are the same.
func packagesNamedAfter(parent, enum string) string {
	var packages []string
	for _, name := range []string{toCamelCase(enum), toLowerSnakeCase(enum)} {
		if pkg := strconv.Quote(qualifiedName(parent, name)); !slices.Contains(packages, pkg) {
			packages = append(packages, pkg)
		}
	}
	return strings.Join(packages, " or ")
}
