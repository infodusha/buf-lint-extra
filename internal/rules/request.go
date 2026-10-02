package rules

import (
	optionv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/option/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type field struct {
	number   protowire.Number
	wireType protowire.Type
}

var (
	checkRequestFileDescriptors = field{1, protowire.BytesType}
	checkRequestOptions         = field{3, protowire.BytesType}
	checkRequestRuleIDs         = field{4, protowire.BytesType}

	fileDescriptorFileDescriptorProto = field{1, protowire.BytesType}
	fileDescriptorIsImport            = field{2, protowire.VarintType}

	fileDescriptorProtoName        = field{1, protowire.BytesType}
	fileDescriptorProtoMessageType = field{4, protowire.BytesType}
	fileDescriptorProtoEnumType    = field{5, protowire.BytesType}
	fileDescriptorProtoService     = field{6, protowire.BytesType}
	fileDescriptorProtoExtension   = field{7, protowire.BytesType}

	enumDescriptorProtoName = field{1, protowire.BytesType}
)

type checkRequest struct {
	files   []fileSummary
	options []*optionv1.Option
	ruleIDs []string
}

func parseCheckRequest(data []byte) (*checkRequest, error) {
	request := &checkRequest{}
	err := rangeFields(data, func(f field, value []byte) error {
		switch f {
		case checkRequestFileDescriptors:
			file, isImport, err := parseFileDescriptor(value)
			if err != nil {
				return err
			}
			if !isImport {
				request.files = append(request.files, file)
			}
		case checkRequestOptions:
			protoOption := &optionv1.Option{}
			if err := proto.Unmarshal(value, protoOption); err != nil {
				return err
			}
			request.options = append(request.options, protoOption)
		case checkRequestRuleIDs:
			request.ruleIDs = append(request.ruleIDs, string(value))
		}
		return nil
	})
	return request, err
}

func parseFileDescriptor(data []byte) (file fileSummary, isImport bool, err error) {
	var fileDescriptorProto []byte
	err = rangeFields(data, func(f field, value []byte) error {
		switch f {
		case fileDescriptorFileDescriptorProto:
			fileDescriptorProto = value
		case fileDescriptorIsImport:
			v, _ := protowire.ConsumeVarint(value)
			isImport = v != 0
		}
		return nil
	})
	if err != nil || isImport {
		return fileSummary{}, isImport, err
	}
	file, err = parseFileDescriptorProto(fileDescriptorProto)
	return file, false, err
}

func parseFileDescriptorProto(data []byte) (fileSummary, error) {
	var file fileSummary
	err := rangeFields(data, func(f field, value []byte) error {
		switch f {
		case fileDescriptorProtoName:
			file.name = string(value)
		case fileDescriptorProtoMessageType:
			file.messages++
		case fileDescriptorProtoEnumType:
			name, err := parseEnumDescriptorProtoName(value)
			if err != nil {
				return err
			}
			file.enums = append(file.enums, name)
		case fileDescriptorProtoService:
			file.services++
		case fileDescriptorProtoExtension:
			file.extensions++
		}
		return nil
	})
	return file, err
}

func parseEnumDescriptorProtoName(data []byte) (string, error) {
	var name string
	err := rangeFields(data, func(f field, value []byte) error {
		if f == enumDescriptorProtoName {
			name = string(value)
		}
		return nil
	})
	return name, err
}

func rangeFields(data []byte, fn func(f field, value []byte) error) error {
	for len(data) > 0 {
		number, wireType, n := protowire.ConsumeTag(data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		data = data[n:]
		n = protowire.ConsumeFieldValue(number, wireType, data)
		if n < 0 {
			return protowire.ParseError(n)
		}
		value := data[:n]
		data = data[n:]
		if wireType == protowire.BytesType {
			value, _ = protowire.ConsumeBytes(value)
		}
		if err := fn(field{number, wireType}, value); err != nil {
			return err
		}
	}
	return nil
}
