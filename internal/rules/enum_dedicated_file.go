package rules

import (
	"context"
	"fmt"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const EnumDedicatedFileRuleID = "ENUM_DEDICATED_FILE"

var enumDedicatedFileRuleSpec = &check.RuleSpec{
	ID:      EnumDedicatedFileRuleID,
	Default: true,
	Purpose: "Checks that top-level enums are declared in dedicated files that contain no messages, services, or extensions.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewFileRuleHandler(checkEnumDedicatedFile, checkutil.WithoutImports()),
}

func checkEnumDedicatedFile(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fileDescriptor descriptor.FileDescriptor,
) error {
	file := fileDescriptor.ProtoreflectFileDescriptor()
	enums := file.Enums()
	if enums.Len() == 0 {
		return nil
	}
	others := nonEnumDeclarations(file)
	if others == "" {
		return nil
	}
	for i := range enums.Len() {
		enum := enums.Get(i)
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Enum %q must be declared in a dedicated file that contains only enums, but this file also declares %s.",
				enum.Name(),
				others,
			),
			check.WithDescriptor(enum),
		)
	}
	return nil
}

// nonEnumDeclarations returns a summary such as "2 messages and 1 service",
// or "" if the file declares nothing but enums.
func nonEnumDeclarations(file protoreflect.FileDescriptor) string {
	var parts []string
	for _, kind := range []struct {
		count int
		noun  string
	}{
		{file.Messages().Len(), "message"},
		{file.Services().Len(), "service"},
		{file.Extensions().Len(), "extension"},
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
