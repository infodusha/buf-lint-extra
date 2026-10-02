package rules

import (
	"context"

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
	name       string
	enums      []string
	messages   int
	services   int
	extensions int
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
		enums:      make([]string, enums.Len()),
		messages:   fileDescriptor.Messages().Len(),
		services:   fileDescriptor.Services().Len(),
		extensions: fileDescriptor.Extensions().Len(),
	}
	for i := range enums.Len() {
		file.enums[i] = string(enums.Get(i).Name())
	}
	return file
}
