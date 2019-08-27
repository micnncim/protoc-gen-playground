package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	errors "golang.org/x/xerrors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	registry, err := makeFilesRegistry(req)
	if err != nil {
		return err
	}
	registry.RangeFiles(func(desc protoreflect.FileDescriptor) bool {
		fmt.Println(desc.FullName())
		return true
	})
	return nil
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
