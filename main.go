package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	errors "golang.org/x/xerrors"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	req, err := parseReq(os.Stdin)
	if err != nil {
		return err
	}
	resp, err := processReq(req)
	if err != nil {
		return err
	}
	return emitResp(resp)
}

func parseReq(r io.Reader) (*pluginpb.CodeGeneratorRequest, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(buf, req); err != nil {
		return nil, errors.Errorf("proto.Unmarshal error: %v", err)
	}
	return req, nil
}

func emitResp(resp *pluginpb.CodeGeneratorResponse) error {
	buf, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(buf); err != nil {
		return err
	}
	return nil
}

func processReq(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	files := make([]*pluginpb.CodeGeneratorResponse_File, 0, len(req.ProtoFile))
	for _, f := range req.ProtoFile {
		content, err := prototext.MarshalOptions{
			Indent: "  ",
		}.Marshal(f)
		if err != nil {
			return nil, err
		}
		files = append(files, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(*f.Name + ".dump"),
			Content: proto.String(string(content)),
		})
	}

	return &pluginpb.CodeGeneratorResponse{
		File: files,
	}, nil
}

func makeFilesRegistry(req *pluginpb.CodeGeneratorRequest) (*protoregistry.Files, error) {
	registry := protoregistry.NewFiles()
	for _, f := range req.GetProtoFile() {
		file, err := protodesc.NewFile(f, registry)
		if err != nil {
			return nil, errors.Errorf("protodesc.NewFile error: %v", err)
		}
		if err := registry.Register(file); err != nil {
			return nil, errors.Errorf("registry.Register error: %v", err)
		}
	}
	return registry, nil
}
