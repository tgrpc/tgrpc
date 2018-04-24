package tgrpc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

var (
	lock sync.Mutex
)

// protoPath 与 incImp 不能有重叠部分 incImp例子：服务目录/服务proto文件
func GenDescriptorSet(protoPath, descSetOut, incImp string) error {
	if protoPath[0] != "/"[0] {
		var prefix string
		if protoPath[0] == "$"[0] {
			spl := strings.SplitN(protoPath, "/", 2)
			prefix = os.Getenv(spl[0][1:])
			protoPath = spl[1]
		} else {
			var err error
			prefix, err = os.Getwd()
			if isErr(err) {
				return err
			}
		}
		protoPath = filepath.Join(prefix, protoPath)
	}
	incImp = getServiceProto(incImp)
	args := []string{fmt.Sprintf("--proto_path=%s", protoPath), fmt.Sprintf("--descriptor_set_out=%s", descSetOut), "--include_imports", fmt.Sprintf("%s", incImp)}

	lock.Lock()
	bs, err := exec.Command("protoc", args...).CombinedOutput()
	lock.Unlock()
	if len(bs) > 0 {
		log.WithField("protoc", string(bs)).Error(err)
	}
	return err
}

// method: pkg.Service incImp:pkg.service.proto
func GetDescriptor(protoBasePath, method, incImp string, reuseDesc bool) (*descriptor.FileDescriptorSet, error) {
	serviceName, err := getServiceName(method)
	if isErr(err) {
		return nil, err
	}
	descSetOut := "." + serviceName + ".pbin"

	var desc *descriptor.FileDescriptorSet

	if reuseDesc {
		desc, err := decodeDesc(descSetOut)
		if err == nil {
			log.WithField("FileDescriptorSet", descSetOut).Debug("use exist desc")
			return desc, nil
		}
	}

	err = GenDescriptorSet(protoBasePath, descSetOut, incImp)
	if isErr(err) {
		return nil, err
	}
	desc, err = decodeDesc(descSetOut)
	if isErr(err) {
		return nil, err
	}
	return desc, nil
}

func getServiceProto(protoFile string) string {
	spl := strings.Split(protoFile, "/")
	size := len(spl)
	if size <= 2 {
		return protoFile
	}
	return strings.Join(spl[size-2:], "/")
}

// helloworld.Greeter
func getServiceName(method string) (string, error) {
	spl := strings.Split(method, "/")
	size := len(spl)
	if size < 2 {
		return "", fmt.Errorf("invalid gRPC method: %s", method)
	}
	return spl[size-2], nil
}

// exp: helloworld.Greeter/SayHello
func getMethod(method string) (string, error) {
	split := strings.Split(method, "/")
	if len(split) < 2 {
		return "", fmt.Errorf("invalid gRPC method: %s", method)
	}
	return strings.Join(split[len(split)-2:], "/"), nil
}

func decodeDesc(descriptorSetFilePath string) (*descriptor.FileDescriptorSet, error) {
	data, err := ioutil.ReadFile(descriptorSetFilePath)
	if err != nil {
		return nil, err
	}
	fileDescriptorSet := &descriptor.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fileDescriptorSet); err != nil {
		return nil, err
	}
	return fileDescriptorSet, nil
}

// Descriptor is an extracted service.
type ServiceDescriptor struct {
	*descriptor.ServiceDescriptorProto

	FullyQualifiedPath  string
	FileDescriptorProto *descriptor.FileDescriptorProto
	FileDescriptorSet   *descriptor.FileDescriptorSet
}

func GetServiceDescriptor(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*ServiceDescriptor, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if path[0] == '.' {
		path = path[1:]
	}

	var serviceDescriptorProto *descriptor.ServiceDescriptorProto
	var fileDescriptorProto *descriptor.FileDescriptorProto
	var fileDescriptorSet *descriptor.FileDescriptorSet
	for _, iFileDescriptorSet := range fileDescriptorSets {
		for _, iFileDescriptorProto := range iFileDescriptorSet.File {
			iServiceDescriptorProto, err := findServiceDescriptorProto(path, iFileDescriptorProto)
			if err != nil {
				return nil, err
			}
			if iServiceDescriptorProto != nil {
				if serviceDescriptorProto != nil {
					return nil, fmt.Errorf("duplicate services for path %s", path)
				}
				serviceDescriptorProto = iServiceDescriptorProto
				fileDescriptorProto = iFileDescriptorProto
			}
		}
		// return first fileDescriptorSet that matches
		// as opposed to duplicate check within fileDescriptorSet, we easily could
		// have multiple fileDescriptorSets that match
		if serviceDescriptorProto != nil {
			fileDescriptorSet = iFileDescriptorSet
			break
		}
	}
	if serviceDescriptorProto == nil {
		return nil, fmt.Errorf("no service for path %s", path)
	}
	return &ServiceDescriptor{
		ServiceDescriptorProto: serviceDescriptorProto,
		FullyQualifiedPath:     "." + path,
		FileDescriptorProto:    fileDescriptorProto,
		FileDescriptorSet:      fileDescriptorSet,
	}, nil
}

func SortFileDescriptorSet(fileDescriptorSet *descriptor.FileDescriptorSet, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.FileDescriptorSet, error) {
	// best-effort checks
	names := make(map[string]struct{}, len(fileDescriptorSet.File))
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() == "" {
			return nil, fmt.Errorf("no name on FileDescriptorProto")
		}
		if _, ok := names[iFileDescriptorProto.GetName()]; ok {
			return nil, fmt.Errorf("duplicate FileDescriptorProto in FileDescriptorSet: %s", iFileDescriptorProto.GetName())
		}
		names[iFileDescriptorProto.GetName()] = struct{}{}
	}
	if _, ok := names[fileDescriptorProto.GetName()]; !ok {
		return nil, fmt.Errorf("no FileDescriptorProto named %s in FileDescriptorSet with names %v", fileDescriptorProto.GetName(), names)
	}
	newFileDescriptorSet := &descriptor.FileDescriptorSet{}
	for _, iFileDescriptorProto := range fileDescriptorSet.File {
		if iFileDescriptorProto.GetName() != fileDescriptorProto.GetName() {
			newFileDescriptorSet.File = append(newFileDescriptorSet.File, iFileDescriptorProto)
		}
	}
	newFileDescriptorSet.File = append(newFileDescriptorSet.File, fileDescriptorProto)
	return newFileDescriptorSet, nil
}

// TODO: we don't actually do full path resolution per the descriptor.proto spec
// https://github.com/google/protobuf/blob/master/src/google/protobuf/descriptor.proto#L185
func findDescriptorProto(path string, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.DescriptorProto, error) {
	if fileDescriptorProto.GetPackage() == "" {
		return nil, fmt.Errorf("no package on FileDescriptorProto")
	}
	if !strings.HasPrefix(path, fileDescriptorProto.GetPackage()) {
		return nil, nil
	}
	return findDescriptorProtoInSlice(path, fileDescriptorProto.GetPackage(), fileDescriptorProto.GetMessageType())
}

func findDescriptorProtoInSlice(path string, nestedName string, descriptorProtos []*descriptor.DescriptorProto) (*descriptor.DescriptorProto, error) {
	var foundDescriptorProto *descriptor.DescriptorProto
	for _, descriptorProto := range descriptorProtos {
		if descriptorProto.GetName() == "" {
			return nil, fmt.Errorf("no name on DescriptorProto")
		}
		fullName := nestedName + "." + descriptorProto.GetName()
		if path == fullName {
			if foundDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate messages for path %s", path)
			}
			foundDescriptorProto = descriptorProto
		}
		nestedFoundDescriptorProto, err := findDescriptorProtoInSlice(path, fullName, descriptorProto.GetNestedType())
		if err != nil {
			return nil, err
		}
		if nestedFoundDescriptorProto != nil {
			if foundDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate messages for path %s", path)
			}
			foundDescriptorProto = nestedFoundDescriptorProto
		}
	}
	return foundDescriptorProto, nil
}

func findServiceDescriptorProto(path string, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.ServiceDescriptorProto, error) {
	if fileDescriptorProto.GetPackage() == "" {
		return nil, fmt.Errorf("no package on FileDescriptorProto")
	}
	if !strings.HasPrefix(path, fileDescriptorProto.GetPackage()) {
		return nil, nil
	}
	var foundServiceDescriptorProto *descriptor.ServiceDescriptorProto
	for _, serviceDescriptorProto := range fileDescriptorProto.GetService() {
		if fileDescriptorProto.GetPackage()+"."+serviceDescriptorProto.GetName() == path {
			if foundServiceDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate services for path %s", path)
			}
			foundServiceDescriptorProto = serviceDescriptorProto
		}
	}
	return foundServiceDescriptorProto, nil
}
