package rules

import (
	"fmt"
	"path"
	"strings"

	"buf.build/go/bufplugin/check"
)

const EnumFileMatchRuleID = "ENUM_FILE_MATCH"

var enumFileMatchRule = newRule(
	&check.RuleSpec{
		ID:      EnumFileMatchRuleID,
		Default: false,
		Purpose: "Checks that files declaring only enums are named after the enum, in any case style and with or without the enum file suffix, so that every enum can be found by its file name.",
		Type:    check.RuleTypeLint,
	},
	checkEnumFileMatch,
)

func checkEnumFileMatch(file fileSummary, request checkRequest) ([]annotation, error) {
	suffix, err := enumFileSuffix(request.options)
	if err != nil {
		return nil, err
	}
	if !file.declaresOnlyEnums() {
		return nil, nil
	}
	stem := strings.TrimSuffix(path.Base(file.name), protoFileExtension)
	name, hasSuffix := strings.CutSuffix(stem, suffix)
	var annotations []annotation
	for i, enum := range file.enums {
		if sameWords(enum, stem) || (hasSuffix && sameWords(enum, name)) {
			continue
		}
		var files []string
		for _, converted := range []string{toLowerKebabCase(enum), toLowerSnakeCase(enum)} {
			if hasSuffix {
				converted += suffix
			}
			files = append(files, path.Join(path.Dir(file.name), converted+protoFileExtension))
		}
		annotations = append(annotations, annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared in a file named after it, such as %s, but the file is %q.",
				enum,
				quotedAlternatives(files),
				file.name,
			),
			sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
		})
	}
	return annotations, nil
}
