package rules

import (
	"fmt"
	"path"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const (
	EnumFileSuffixRuleID = "ENUM_FILE_SUFFIX"

	// The option value is accepted with or without the ".proto" extension.
	EnumFileSuffixOptionKey = "enum_file_suffix"

	DefaultEnumFileSuffix = "_enum"

	protoFileExtension = ".proto"
)

var enumFileSuffixRule = newRule(
	&check.RuleSpec{
		ID:      EnumFileSuffixRuleID,
		Default: false,
		Purpose: `Checks that files declaring top-level enums have a name ending in a specific suffix (default is "_enum"), and that files with that suffix declare top-level enums.`,
		Type:    check.RuleTypeLint,
	},
	checkEnumFileSuffix,
)

func checkEnumFileSuffix(file fileSummary, options option.Options) ([]annotation, error) {
	suffix, err := enumFileSuffix(options)
	if err != nil {
		return nil, err
	}
	stem := strings.TrimSuffix(path.Base(file.name), protoFileExtension)
	hasSuffix := strings.HasSuffix(stem, suffix)
	hasEnums := len(file.enums) > 0
	others := nonEnumDeclarations(file)
	var message string
	switch {
	case hasEnums && !hasSuffix && others != "":
		message = fmt.Sprintf(
			"File %q declares top-level enums alongside %s, so the enums must move to a file with a name ending in %q.",
			file.name,
			others,
			suffix+protoFileExtension,
		)
	case hasEnums && !hasSuffix:
		message = fmt.Sprintf(
			"File %q declares top-level enums and must have a name ending in %q, such as %q.",
			file.name,
			suffix+protoFileExtension,
			path.Join(path.Dir(file.name), stem+suffix+protoFileExtension),
		)
	case hasSuffix && !hasEnums:
		message = fmt.Sprintf(
			"File %q has a name ending in %q but declares no top-level enums.",
			file.name,
			suffix+protoFileExtension,
		)
	default:
		return nil, nil
	}
	return []annotation{{message: message}}, nil
}

func enumFileSuffix(options option.Options) (string, error) {
	value, err := option.GetStringValue(options, EnumFileSuffixOptionKey)
	if err != nil {
		return "", err
	}
	if value == "" {
		return DefaultEnumFileSuffix, nil
	}
	suffix := strings.TrimSuffix(value, protoFileExtension)
	if suffix == "" {
		return "", fmt.Errorf("option %q must contain a suffix in addition to %q, got %q", EnumFileSuffixOptionKey, protoFileExtension, value)
	}
	return suffix, nil
}
