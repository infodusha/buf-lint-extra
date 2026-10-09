package rules

import (
	"context"
	"slices"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type rule struct {
	spec  *check.RuleSpec
	check checkFunc
}

type checkFunc func(file fileSummary, options option.Options) ([]annotation, error)

type fileSummary struct {
	name        string
	pkg         string
	enums       []string
	nestedEnums []nestedEnum
	messages    int
	services    int
	extensions  int
}

type nestedEnum struct {
	name       string
	message    string
	sourcePath []int32
}

type annotation struct {
	message    string
	sourcePath []int32
}

func newRule(spec *check.RuleSpec, checkFile checkFunc) rule {
	spec.Handler = checkutil.NewFileRuleHandler(
		func(
			_ context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			fileDescriptor descriptor.FileDescriptor,
		) error {
			file := fileSummaryForDescriptor(fileDescriptor.ProtoreflectFileDescriptor())
			annotations, err := checkFile(file, request.Options())
			if err != nil {
				return err
			}
			for _, annotation := range annotations {
				responseWriter.AddAnnotation(
					check.WithMessage(annotation.message),
					check.WithFileNameAndSourcePath(file.name, annotation.sourcePath),
				)
			}
			return nil
		},
		checkutil.WithoutImports(),
	)
	return rule{spec: spec, check: checkFile}
}

func fileSummaryForDescriptor(fileDescriptor protoreflect.FileDescriptor) fileSummary {
	enums := fileDescriptor.Enums()
	file := fileSummary{
		name:       fileDescriptor.Path(),
		pkg:        string(fileDescriptor.Package()),
		enums:      make([]string, enums.Len()),
		messages:   fileDescriptor.Messages().Len(),
		services:   fileDescriptor.Services().Len(),
		extensions: fileDescriptor.Extensions().Len(),
	}
	for i := range enums.Len() {
		file.enums[i] = string(enums.Get(i).Name())
	}
	messages := fileDescriptor.Messages()
	for i := range messages.Len() {
		file.nestedEnums = appendNestedEnums(
			file.nestedEnums,
			messages.Get(i),
			"",
			[]int32{int32(fileDescriptorProtoMessageType.number), int32(i)},
		)
	}
	return file
}

func appendNestedEnums(
	nestedEnums []nestedEnum,
	message protoreflect.MessageDescriptor,
	scope string,
	sourcePath []int32,
) []nestedEnum {
	name := qualifiedName(scope, string(message.Name()))
	enums := message.Enums()
	for i := range enums.Len() {
		nestedEnums = append(nestedEnums, nestedEnum{
			name:       qualifiedName(name, string(enums.Get(i).Name())),
			message:    name,
			sourcePath: childSourcePath(sourcePath, descriptorProtoEnumType, i),
		})
	}
	messages := message.Messages()
	for i := range messages.Len() {
		nestedEnums = appendNestedEnums(
			nestedEnums,
			messages.Get(i),
			name,
			childSourcePath(sourcePath, descriptorProtoNestedType, i),
		)
	}
	return nestedEnums
}

func qualifiedName(scope, name string) string {
	if scope == "" {
		return name
	}
	return scope + "." + name
}

func childSourcePath(parent []int32, f field, index int) []int32 {
	return append(slices.Clip(parent), int32(f.number), int32(index))
}
