package rules

import (
	"context"
	"fmt"
	"path"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/option"
)

const (
	EnumFileSuffixRuleID = "ENUM_FILE_SUFFIX"

	// The option value is accepted with or without the ".proto" extension.
	EnumFileSuffixOptionKey = "enum_file_suffix"

	DefaultEnumFileSuffix = "_enum"

	protoFileExtension = ".proto"
)

var enumFileSuffixRuleSpec = &check.RuleSpec{
	ID:      EnumFileSuffixRuleID,
	Default: false,
	Purpose: `Checks that files declaring top-level enums have a name ending in a specific suffix (default is "_enum"), and that files with that suffix declare top-level enums.`,
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewFileRuleHandler(checkEnumFileSuffix, checkutil.WithoutImports()),
}

func checkEnumFileSuffix(
	_ context.Context,
	responseWriter check.ResponseWriter,
	request check.Request,
	fileDescriptor descriptor.FileDescriptor,
) error {
	suffix, err := enumFileSuffix(request.Options())
	if err != nil {
		return err
	}
	file := fileDescriptor.ProtoreflectFileDescriptor()
	fileName := file.Path()
	stem := strings.TrimSuffix(path.Base(fileName), protoFileExtension)
	hasSuffix := strings.HasSuffix(stem, suffix)
	hasEnums := file.Enums().Len() > 0
	others := nonEnumDeclarations(file)
	switch {
	case hasEnums && !hasSuffix && others != "":
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"File %q declares top-level enums alongside %s, so the enums must move to a file with a name ending in %q.",
				fileName,
				others,
				suffix+protoFileExtension,
			),
			check.WithFileName(fileName),
		)
	case hasEnums && !hasSuffix:
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"File %q declares top-level enums and must have a name ending in %q, such as %q.",
				fileName,
				suffix+protoFileExtension,
				path.Join(path.Dir(fileName), stem+suffix+protoFileExtension),
			),
			check.WithFileName(fileName),
		)
	case hasSuffix && !hasEnums:
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"File %q has a name ending in %q but declares no top-level enums.",
				fileName,
				suffix+protoFileExtension,
			),
			check.WithFileName(fileName),
		)
	}
	return nil
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
