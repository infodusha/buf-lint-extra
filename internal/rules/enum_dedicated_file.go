package rules

import (
	"fmt"
	"strings"

	"buf.build/go/bufplugin/check"
)

const EnumDedicatedFileRuleID = "ENUM_DEDICATED_FILE"

var enumDedicatedFileRule = newRule(
	&check.RuleSpec{
		ID:      EnumDedicatedFileRuleID,
		Default: true,
		Purpose: "Checks that each enum is declared at the top level of a file of its own, with no other enums, messages, services, or extensions, and not nested in a message.",
		Type:    check.RuleTypeLint,
	},
	checkEnumDedicatedFile,
)

func checkEnumDedicatedFile(file fileSummary, _ checkRequest) ([]annotation, error) {
	var annotations []annotation
	parts := nonEnumDeclarationParts(file)
	if n := len(file.enums) - 1; n > 0 {
		parts = append([]string{countNoun(n, "other enum")}, parts...)
	}
	if others := joinWithAnd(parts); others != "" {
		for i, enum := range file.enums {
			annotations = append(annotations, annotation{
				message: fmt.Sprintf(
					"Enum %q must be declared in a file of its own, but this file also declares %s.",
					enum,
					others,
				),
				sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
			})
		}
	}
	for _, enum := range file.nestedEnums {
		annotations = append(annotations, annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared at the top level of a file of its own, not nested in message %q.",
				enum.name,
				enum.message,
			),
			sourcePath: enum.sourcePath,
		})
	}
	return annotations, nil
}

func nonEnumDeclarationParts(file fileSummary) []string {
	var parts []string
	for _, kind := range []struct {
		count int
		noun  string
	}{
		{file.messages, "message"},
		{file.services, "service"},
		{file.extensions, "extension"},
	} {
		if kind.count > 0 {
			parts = append(parts, countNoun(kind.count, kind.noun))
		}
	}
	return parts
}

func countNoun(count int, noun string) string {
	if count != 1 {
		noun += "s"
	}
	return fmt.Sprintf("%d %s", count, noun)
}

func joinWithAnd(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
	}
}
