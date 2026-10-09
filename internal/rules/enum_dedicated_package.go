package rules

import (
	"fmt"

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

func checkEnumDedicatedPackage(file fileSummary, request checkRequest) ([]annotation, error) {
	if file.pkg == "" || !file.declaresOnlyEnums() {
		return nil, nil
	}
	parent, last := splitLastComponent(file.pkg)
	cases := []func(string) string{toCamelCase, toLowerSnakeCase}
	if request.enables(packageCamelCaseRule) {
		cases = cases[:1]
	}
	var annotations []annotation
	for i, enum := range file.enums {
		if sameWords(enum, last) {
			continue
		}
		var packages []string
		for _, convert := range cases {
			packages = append(packages, qualifiedName(parent, convert(enum)))
		}
		annotations = append(annotations, annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared in a package named after it, such as %s, but the package is %q.",
				enum,
				quotedAlternatives(packages),
				file.pkg,
			),
			sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
		})
	}
	return annotations, nil
}
