package rules

import (
	"fmt"
	"strings"

	"buf.build/go/bufplugin/check"
)

const EnumValueNoPrefixRuleID = "ENUM_VALUE_NO_PREFIX"

var enumValueNoPrefixRule = newRule(
	&check.RuleSpec{
		ID:      EnumValueNoPrefixRuleID,
		Default: false,
		Purpose: "Checks that the values of an enum in a package of its own, the one ENUM_DEDICATED_PACKAGE asks for, are not prefixed with the name of the enum, since there is nothing in the package for them to clash with.",
		Type:    check.RuleTypeLint,
	},
	checkEnumValueNoPrefix,
)

func checkEnumValueNoPrefix(file fileSummary, _ checkRequest) ([]annotation, error) {
	i := file.dedicatedEnum()
	if i < 0 {
		return nil, nil
	}
	enum := file.enums[i]
	prefix := toUpperSnakeCase(enum.name) + "_"
	var annotations []annotation
	for j, value := range enum.values {
		rest, ok := strings.CutPrefix(value.name, prefix)
		if !ok {
			continue
		}
		message := fmt.Sprintf("Enum value name %q should not be prefixed with %q", value.name, prefix)
		if rest != "" {
			message += fmt.Sprintf(", such as %q", rest)
		}
		annotations = append(annotations, annotation{
			message:    message + ".",
			sourcePath: enumValueSourcePath(topLevelEnumSourcePath(i), j),
		})
	}
	return annotations, nil
}
