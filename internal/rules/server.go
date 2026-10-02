package rules

import (
	"context"
	"slices"

	checkv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1"
	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/option"
	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"pluginrpc.com/pluginrpc"
)

const (
	checkPath          = "/buf.plugin.check.v1.CheckService/Check"
	listRulesPath      = "/buf.plugin.check.v1.CheckService/ListRules"
	listCategoriesPath = "/buf.plugin.check.v1.CheckService/ListCategories"
	getPluginInfoPath  = "/buf.plugin.info.v1.PluginInfoService/GetPluginInfo"
)

var rawCheckRequestDescriptor = newRawCheckRequestDescriptor()

func NewServer() (pluginrpc.Server, error) {
	return newServer(noopValidator{})
}

func newServer(validator protovalidate.Validator) (pluginrpc.Server, error) {
	checkServiceHandler, err := check.NewCheckServiceHandler(Spec, check.CheckServiceHandlerWithValidator(validator))
	if err != nil {
		return nil, err
	}
	pluginInfoServiceHandler, err := info.NewPluginInfoServiceHandler(Spec.Info, info.PluginInfoServiceHandlerWithValidator(validator))
	if err != nil {
		return nil, err
	}
	pluginInfo, err := info.NewPluginInfoForSpec(Spec.Info)
	if err != nil {
		return nil, err
	}
	var procedures []pluginrpc.Procedure
	for _, procedure := range []struct {
		path string
		args string
	}{
		{checkPath, "check"},
		{listRulesPath, "list-rules"},
		{listCategoriesPath, "list-categories"},
		{getPluginInfoPath, "info"},
	} {
		p, err := pluginrpc.NewProcedure(procedure.path, pluginrpc.ProcedureWithArgs(procedure.args))
		if err != nil {
			return nil, err
		}
		procedures = append(procedures, p)
	}
	spec, err := pluginrpc.NewSpec(procedures...)
	if err != nil {
		return nil, err
	}
	handler := pluginrpc.NewHandler(spec)
	serverRegistrar := pluginrpc.NewServerRegistrar()
	serverRegistrar.Register(checkPath, func(ctx context.Context, handleEnv pluginrpc.HandleEnv, options ...pluginrpc.HandleOption) error {
		return handler.Handle(
			ctx,
			handleEnv,
			dynamicpb.NewMessage(rawCheckRequestDescriptor),
			func(_ context.Context, request any) (any, error) {
				return checkRawRequest(validator, request.(*dynamicpb.Message).GetUnknown())
			},
			options...,
		)
	})
	serverRegistrar.Register(listRulesPath, handleFunc(handler, checkServiceHandler.ListRules))
	serverRegistrar.Register(listCategoriesPath, handleFunc(handler, checkServiceHandler.ListCategories))
	serverRegistrar.Register(getPluginInfoPath, handleFunc(handler, pluginInfoServiceHandler.GetPluginInfo))
	return pluginrpc.NewServer(spec, serverRegistrar, pluginrpc.ServerWithDoc(pluginInfo.Documentation()))
}

func handleFunc[Request, Response proto.Message](
	handler pluginrpc.Handler,
	handle func(context.Context, Request) (Response, error),
) func(context.Context, pluginrpc.HandleEnv, ...pluginrpc.HandleOption) error {
	return func(ctx context.Context, handleEnv pluginrpc.HandleEnv, options ...pluginrpc.HandleOption) error {
		var request Request
		return handler.Handle(
			ctx,
			handleEnv,
			request.ProtoReflect().New().Interface(),
			func(ctx context.Context, request any) (any, error) {
				return handle(ctx, request.(Request))
			},
			options...,
		)
	}
}

func checkRawRequest(validator protovalidate.Validator, data []byte) (*checkv1.CheckResponse, error) {
	request, err := parseCheckRequest(data)
	if err != nil {
		return nil, pluginrpc.NewError(pluginrpc.CodeInvalidArgument, err)
	}
	options, err := option.OptionsForProtoOptions(request.options)
	if err != nil {
		return nil, err
	}
	rules, err := rulesForIDs(request.ruleIDs)
	if err != nil {
		return nil, err
	}
	response := &checkv1.CheckResponse{}
	for _, rule := range rules {
		for _, file := range request.files {
			annotations, err := rule.check(file, options)
			if err != nil {
				return nil, err
			}
			for _, annotation := range annotations {
				response.Annotations = append(response.Annotations, &checkv1.Annotation{
					RuleId:  rule.spec.ID,
					Message: annotation.message,
					FileLocation: &descriptorv1.FileLocation{
						FileName:   file.name,
						SourcePath: annotation.sourcePath,
					},
				})
			}
		}
	}
	if err := validator.Validate(response); err != nil {
		return nil, err
	}
	return response, nil
}

func rulesForIDs(ruleIDs []string) ([]rule, error) {
	if len(ruleIDs) == 0 {
		return slices.DeleteFunc(slices.Clone(allRules), func(r rule) bool { return !r.spec.Default }), nil
	}
	rules := make([]rule, len(ruleIDs))
	for i, ruleID := range ruleIDs {
		index := slices.IndexFunc(allRules, func(r rule) bool { return r.spec.ID == ruleID })
		if index < 0 {
			return nil, pluginrpc.NewErrorf(pluginrpc.CodeInvalidArgument, "unknown rule ID: %q", ruleID)
		}
		rules[i] = allRules[index]
	}
	return rules, nil
}

func newRawCheckRequestDescriptor() protoreflect.MessageDescriptor {
	name := (*checkv1.CheckRequest)(nil).ProtoReflect().Descriptor().FullName()
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:        proto.String("buf_lint_extra/raw_check_request.proto"),
		Package:     proto.String(string(name.Parent())),
		MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String(string(name.Name()))}},
	}, nil)
	if err != nil {
		panic(err)
	}
	return file.Messages().Get(0)
}

type noopValidator struct{}

func (noopValidator) Validate(proto.Message, ...protovalidate.ValidationOption) error {
	return nil
}
