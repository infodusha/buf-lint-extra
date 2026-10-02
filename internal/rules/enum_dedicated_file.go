package rules

import (
	"fmt"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const EnumDedicatedFileRuleID = "ENUM_DEDICATED_FILE"

var enumDedicatedFileRule = newRule(
	&check.RuleSpec{
		ID:      EnumDedicatedFileRuleID,
		Default: true,
		Purpose: "Checks that top-level enums are declared in dedicated files that contain no messages, services, or extensions.",
		Type:    check.RuleTypeLint,
	},
	checkEnumDedicatedFile,
)

func checkEnumDedicatedFile(file fileSummary, _ option.Options) ([]annotation, error) {
	if len(file.enums) == 0 {
		return nil, nil
	}
	others := nonEnumDeclarations(file)
	if others == "" {
		return nil, nil
	}
	annotations := make([]annotation, len(file.enums))
	for i, enum := range file.enums {
		annotations[i] = annotation{
			message: fmt.Sprintf(
				"Enum %q must be declared in a dedicated file that contains only enums, but this file also declares %s.",
				enum,
				others,
			),
			sourcePath: []int32{int32(fileDescriptorProtoEnumType.number), int32(i)},
		}
	}
	return annotations, nil
}

// nonEnumDeclarations returns a summary such as "2 messages and 1 service",
// or "" if the file declares nothing but enums.
func nonEnumDeclarations(file fileSummary) string {
	var parts []string
	for _, kind := range []struct {
		count int
		noun  string
	}{
		{file.messages, "message"},
		{file.services, "service"},
		{file.extensions, "extension"},
	} {
		if kind.count == 0 {
			continue
		}
		noun := kind.noun
		if kind.count != 1 {
			noun += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", kind.count, noun))
	}
	return joinWithAnd(parts)
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
