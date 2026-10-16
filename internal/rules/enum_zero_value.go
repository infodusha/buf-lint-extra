package rules

import (
	"fmt"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const (
	EnumZeroValueRuleID = "ENUM_ZERO_VALUE"

	EnumZeroValueOptionKey = "enum_zero_value"

	DefaultEnumZeroValue = "UNSPECIFIED"
)

var enumZeroValueRule = newRule(
	&check.RuleSpec{
		ID:      EnumZeroValueRuleID,
		Default: false,
		Purpose: `Checks that enum zero values have a consistent name (configurable, default name is "UNSPECIFIED"), the whole name rather than a suffix as with the builtin ENUM_ZERO_VALUE_SUFFIX.`,
		Type:    check.RuleTypeLint,
	},
	checkEnumZeroValue,
)

func checkEnumZeroValue(file fileSummary, request checkRequest) ([]annotation, error) {
	name, err := enumZeroValue(request.options)
	if err != nil {
		return nil, err
	}
	var annotations []annotation
	for i, enum := range file.enums {
		annotations = append(annotations, enumZeroValueAnnotations(enum, topLevelEnumSourcePath(i), name)...)
	}
	for _, enum := range file.nestedEnums {
		annotations = append(annotations, enumZeroValueAnnotations(enum.enumSummary, enum.sourcePath, name)...)
	}
	return annotations, nil
}

func enumZeroValueAnnotations(enum enumSummary, sourcePath []int32, name string) []annotation {
	var annotations []annotation
	for j, value := range enum.values {
		if value.number != 0 || value.name == name {
			continue
		}
		annotations = append(annotations, annotation{
			message:    fmt.Sprintf("Enum zero value name %q should be %q.", value.name, name),
			sourcePath: enumValueSourcePath(sourcePath, j),
		})
	}
	return annotations
}

func enumZeroValue(options option.Options) (string, error) {
	value, err := option.GetStringValue(options, EnumZeroValueOptionKey)
	if err != nil {
		return "", err
	}
	if value == "" {
		return DefaultEnumZeroValue, nil
	}
	return value, nil
}
