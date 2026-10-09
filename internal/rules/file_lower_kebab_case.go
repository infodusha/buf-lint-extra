package rules

import (
	"fmt"
	"path"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const FileLowerKebabCaseRuleID = "FILE_LOWER_KEBAB_CASE"

var fileLowerKebabCaseRule = newRule(
	&check.RuleSpec{
		ID:      FileLowerKebabCaseRuleID,
		Default: false,
		Purpose: "Checks that each dot-separated segment of a filename is lower-kebab-case, with words separated by hyphens.",
		Type:    check.RuleTypeLint,
	},
	checkFileLowerKebabCase,
)

func checkFileLowerKebabCase(file fileSummary, _ option.Options) ([]annotation, error) {
	base := path.Base(file.name)
	stem := strings.TrimSuffix(base, protoFileExtension)
	var segments []string
	for segment := range strings.SplitSeq(stem, ".") {
		if segment = toLowerKebabCase(segment); segment != "" {
			segments = append(segments, segment)
		}
	}
	expected := strings.Join(segments, ".")
	if expected == stem {
		return nil, nil
	}
	return []annotation{{
		message: fmt.Sprintf(
			"Filename %q should be lower-kebab-case%s, such as %q.",
			base,
			protoFileExtension,
			expected+protoFileExtension,
		),
	}}, nil
}

// toLowerKebabCase joins the words of s with hyphens in lowercase, so
// "user_service", "UserService" and "HTTPServer" become "user-service" and
// "http-server".
func toLowerKebabCase(s string) string {
	return strings.ToLower(strings.Join(splitWords(s), "-"))
}
