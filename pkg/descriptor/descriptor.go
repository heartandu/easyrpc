package descriptor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/heartandu/easyrpc/pkg/fs"
)

// ErrSymbolNotFound is returned when a symbol is not found in the protocol buffer files.
var ErrSymbolNotFound = errors.New("symbol not found")

// ErrReflectionNotSupported is returned when the server does not support the reflection API.
var ErrReflectionNotSupported = errors.New("server does not support reflection API")

// Source defines the interface for a source of protocol buffer descriptors.
type Source interface {
	ListServices() ([]string, error)
	ListMethods() ([]string, error)
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
			Accessor: func(path string) (io.ReadCloser, error) {
				p, err := fs.ExpandHome(path)
				if err != nil {
					return nil, fmt.Errorf("failed to expand home: %w", err)
				}

				return os.Open(p)
			},
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
	services := make([]string, 0)

	for service := range s.services() {
		services = append(services, string(service.FullName()))
	}

	return services, nil
}

// ListMethods returns a list of methods in the protocol buffer files.
func (s *protoFilesSource) ListMethods() ([]string, error) {
	methods := make([]string, 0)

	for service := range s.services() {
		m := service.Methods()
		for i := range m.Len() {
			methods = append(methods, string(m.Get(i).FullName()))
		}
	}

	return methods, nil
}

func (s *protoFilesSource) services() iter.Seq[protoreflect.ServiceDescriptor] {
	return func(yield func(protoreflect.ServiceDescriptor) bool) {
		for _, fd := range s.fds {
			for i := range fd.Services().Len() {
				if cont := yield(fd.Services().Get(i)); !cont {
					return
				}
			}
		}
	}
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
	c *grpcreflect.Client
}

// ListServices returns a list of services using server reflection.
func (s *serverReflectionSource) ListServices() ([]string, error) {
	services, err := s.c.ListServices()
	if err != nil {
		return nil, reflectWrapErr("failed to query service list", err)
	}

	return services, nil
}

// ListMethods returns a list of methods using server reflection.
func (s *serverReflectionSource) ListMethods() ([]string, error) {
	services, err := s.c.ListServices()
	if err != nil {
		return nil, reflectWrapErr("failed to query service list", err)
	}

	methods := make([]string, 0)

	for _, service := range services {
		fd, err := s.c.FileContainingSymbol(service)
		if err != nil {
			return nil, reflectWrapErr("failed to query file containing symbol", err)
		}

		for _, fds := range fd.GetServices() {
			for _, md := range fds.GetMethods() {
				methods = append(methods, md.GetFullyQualifiedName())
			}
		}
	}

	return methods, nil
}

// FindSymbol finds a symbol using server reflection.
func (s *serverReflectionSource) FindSymbol(name string) (protoreflect.Descriptor, error) {
	fileDescriptor, err := s.c.FileContainingSymbol(name)
	if err != nil {
		return nil, reflectWrapErr("failed to query file containing symbol", err)
	}

	if d := fileDescriptor.FindSymbol(name); d != nil {
		if wr, ok := d.(interface {
			Unwrap() protoreflect.Descriptor
		}); ok {
			return wr.Unwrap(), nil
		}
	}

	return nil, ErrSymbolNotFound
}

func reflectWrapErr(msg string, err error) error {
	if err == nil {
		return nil
	}

	if st, ok := status.FromError(err); ok && st.Code() == codes.Unimplemented {
		return ErrReflectionNotSupported
	}

	return fmt.Errorf("%s: %w", msg, err)
}
