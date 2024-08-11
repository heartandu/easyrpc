package descriptor

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ErrSymbolNotFound is returned when a symbol is not found in the protocol buffer files.
var ErrSymbolNotFound = errors.New("symbol not found")

// ErrReflectionNotSupported is returned when the server does not support the reflection API.
var ErrReflectionNotSupported = errors.New("server does not support reflection API")

// Source defines the interface for a source of protocol buffer descriptors.
type Source interface {
	ListServices() ([]string, error)
	FindSymbol(name string) (protoreflect.Descriptor, error)
}

// ReflectionSource creates a source of protocol buffer descriptors using server reflection.
func ReflectionSource(ctx context.Context, cc grpc.ClientConnInterface) (Source, error) {
	return &serverReflectionSource{c: grpcreflect.NewClientAuto(ctx, cc)}, nil
}

// ProtoFilesSource creates a source of protocol buffer descriptors using proto files.
func ProtoFilesSource(ctx context.Context, importPaths, protoFiles []string) (Source, error) {
	comp := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: importPaths,
		}),
	}

	fds, err := comp.Compile(ctx, protoFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto files: %w", err)
	}

	return &protoFilesSource{
		fds: fds,
	}, nil
}

type protoFilesSource struct {
	fds linker.Files
}

// ListServices returns a list of services in the protocol buffer files.
func (s *protoFilesSource) ListServices() ([]string, error) {
	var services []string

	for _, fd := range s.fds {
		for i := range fd.Services().Len() {
			services = append(services, string(fd.Services().Get(i).FullName()))
		}
	}

	return services, nil
}

// FindSymbol finds a symbol in the protocol buffer files.
func (s *protoFilesSource) FindSymbol(name string) (protoreflect.Descriptor, error) {
	for _, fd := range s.fds {
		if d := fd.FindDescriptorByName(protoreflect.FullName(name)); d != nil {
			return d, nil
		}
	}

	return nil, ErrSymbolNotFound
}

type serverReflectionSource struct {
	c interface {
		ListServices() ([]string, error)
		FileContainingSymbol(symbol string) (*desc.FileDescriptor, error)
	}
}

// ListServices returns a list of services using server reflection.
func (s *serverReflectionSource) ListServices() ([]string, error) {
	services, err := s.c.ListServices()

	return services, reflectWrapErr(err)
}

// FindSymbol finds a symbol using server reflection.
func (s *serverReflectionSource) FindSymbol(name string) (protoreflect.Descriptor, error) {
	fd, err := s.c.FileContainingSymbol(name)
	if err != nil {
		return nil, reflectWrapErr(err)
	}

	if d := fd.FindSymbol(name); d != nil {
		if wr, ok := d.(desc.DescriptorWrapper); ok {
			return wr.Unwrap(), nil
		}
	}

	return nil, ErrSymbolNotFound
}

func reflectWrapErr(err error) error {
	if err == nil {
		return nil
	}

	if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
		return ErrReflectionNotSupported
	}

	return err
}
