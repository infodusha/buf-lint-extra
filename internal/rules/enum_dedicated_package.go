package rules

import (
	"fmt"

	"buf.build/go/bufplugin/check"
)

const EnumDedicatedPackageRuleID = "ENUM_DEDICATED_PACKAGE"

var enumDedicatedPackageRule = newFilesRule(
	&check.RuleSpec{
		ID:      EnumDedicatedPackageRuleID,
		Default: false,
		Purpose: "Checks that files declaring only enums have a package whose last component is the name of the enum, in any case style, and that no other file is in that package, so that every enum gets a package of its own.",
		Type:    check.RuleTypeLint,
	},
	checkEnumDedicatedPackage,
)

type dedicatedEnum struct {
	file string
	enum string
}

func checkEnumDedicatedPackage(files []fileSummary, request checkRequest) ([]fileAnnotation, error) {
	dedicatedByPackage := map[string]dedicatedEnum{}
	for _, file := range files {
		if i := file.dedicatedEnum(); i >= 0 {
			dedicatedByPackage[file.pkg] = dedicatedEnum{file: file.name, enum: file.enums[i].name}
		}
	}
	var annotations []fileAnnotation
	for _, file := range files {
		for _, a := range enumDedicatedPackageAnnotations(file, request) {
			annotations = append(annotations, fileAnnotation{fileName: file.name, annotation: a})
		}
		dedicated, ok := dedicatedByPackage[file.pkg]
		if !ok || file.declaresOnlyEnums() {
			continue
		}
		annotations = append(annotations, fileAnnotation{
			fileName: file.name,
			annotation: annotation{
				message: fmt.Sprintf(
					"Package %q is dedicated to enum %q declared in %q, so this file must be in another package.",
					file.pkg,
					dedicated.enum,
					dedicated.file,
				),
				sourcePath: []int32{int32(fileDescriptorProtoPackage.number)},
			},
		})
	}
	return annotations, nil
}

func enumDedicatedPackageAnnotations(file fileSummary, request checkRequest) []annotation {
	if file.pkg == "" || !file.declaresOnlyEnums() {
		return nil
	}
	parent, last := splitLastComponent(file.pkg)
	cases := []func(string) string{toCamelCase, toLowerSnakeCase}
	if request.enables(packageCamelCaseRule) {
		cases = cases[:1]
	}
	var annotations []annotation
	for i, enum := range file.enums {
		if sameWords(enum.name, last) {
			continue
		}
		var packages []string
		for _, convert := range cases {
			packages = append(packages, qualifiedName(parent, convert(enum.name)))
		}
		annotations = append(annotations, annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared in a package named after it, such as %s, but the package is %q.",
				enum.name,
				quotedAlternatives(packages),
				file.pkg,
			),
			sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
		})
	}
	return annotations
}
