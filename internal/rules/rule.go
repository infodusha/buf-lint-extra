package rules

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/proto"
)

type rule struct {
	spec  *check.RuleSpec
	check filesCheckFunc
}

type checkFunc func(file fileSummary, request checkRequest) ([]annotation, error)

type filesCheckFunc func(files []fileSummary, request checkRequest) ([]fileAnnotation, error)

type checkRequest struct {
	options option.Options
	ruleIDs []string
}

// enables reports whether the rule runs in this request: it is requested by
// ID, or the request has no rule IDs and the rule is on by default.
func (request checkRequest) enables(r rule) bool {
	if len(request.ruleIDs) == 0 {
		return r.spec.Default
	}
	return slices.Contains(request.ruleIDs, r.spec.ID)
}

type fileSummary struct {
	name        string
	pkg         string
	enums       []string
	nestedEnums []nestedEnum
	messages    int
	services    int
	extensions  int
}

func (file fileSummary) declaresOnlyEnums() bool {
	return len(file.enums) > 0 && file.messages == 0 && file.services == 0 && file.extensions == 0
}

func (file fileSummary) declaresEnumNamed(name string) bool {
	return slices.ContainsFunc(file.enums, func(enum string) bool { return sameWords(enum, name) })
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

type fileAnnotation struct {
	fileName string
	annotation
}

func newRule(spec *check.RuleSpec, checkFile checkFunc) rule {
	return newFilesRule(spec, func(files []fileSummary, request checkRequest) ([]fileAnnotation, error) {
		var annotations []fileAnnotation
		for _, file := range files {
			fileAnnotations, err := checkFile(file, request)
			if err != nil {
				return nil, err
			}
			for _, a := range fileAnnotations {
				annotations = append(annotations, fileAnnotation{fileName: file.name, annotation: a})
			}
		}
		return annotations, nil
	})
}

func newFilesRule(spec *check.RuleSpec, checkFiles filesCheckFunc) rule {
	spec.CategoryIDs = []string{ExtraCategoryID}
	spec.Handler = check.RuleHandlerFunc(func(_ context.Context, responseWriter check.ResponseWriter, request check.Request) error {
		var files []fileSummary
		for _, fileDescriptor := range request.FileDescriptors() {
			if fileDescriptor.IsImport() {
				continue
			}
			data, err := proto.Marshal(fileDescriptor.FileDescriptorProto())
			if err != nil {
				return err
			}
			file, err := parseFileDescriptorProto(data)
			if err != nil {
				return err
			}
			files = append(files, file)
		}
		annotations, err := checkFiles(files, checkRequest{options: request.Options(), ruleIDs: request.RuleIDs()})
		if err != nil {
			return err
		}
		for _, a := range annotations {
			responseWriter.AddAnnotation(
				check.WithMessage(a.message),
				check.WithFileNameAndSourcePath(a.fileName, a.sourcePath),
			)
		}
		return nil
	})
	return rule{spec: spec, check: checkFiles}
}

func splitLastComponent(pkg string) (parent, last string) {
	if i := strings.LastIndex(pkg, "."); i >= 0 {
		return pkg[:i], pkg[i+1:]
	}
	return "", pkg
}

// quotedAlternatives quotes the distinct values and joins them with "or".
func quotedAlternatives(values []string) string {
	var quoted []string
	for _, value := range values {
		if q := strconv.Quote(value); !slices.Contains(quoted, q) {
			quoted = append(quoted, q)
		}
	}
	return strings.Join(quoted, " or ")
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
