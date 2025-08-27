package proto

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/afero"
	"google.golang.org/grpc"

	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/pkg/descriptor"
)

// NewDescriptorSource returns a new descriptor source based on the provided configuration.
func NewDescriptorSource(
	ctx context.Context,
	fsys afero.Fs,
	cfg *config.Config,
	clientConn grpc.ClientConnInterface,
) (descriptor.Source, error) {
	var (
		descSrc descriptor.Source
		err     error
	)

	if cfg.Server.Reflection {
		descSrc, err = descriptor.ReflectionSource(ctx, clientConn)
	} else {
		importPaths, protoFiles := cfg.Proto.ImportPaths, cfg.Proto.ProtoFiles
		if cfg.Proto.ImportAll {
			protoFiles, err = findProtoFiles(fsys, importPaths)
			if err != nil {
				return nil, fmt.Errorf("failed to find proto files: %w", err)
			}
		}

		descSrc, err = descriptor.ProtoFilesSource(ctx, fsys, importPaths, protoFiles)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create descriptor source: %w", err)
	}

	return descSrc, nil
}

func findProtoFiles(fsys afero.Fs, importPaths []string) ([]string, error) {
	protoFiles := make([]string, 0)

	for _, importPath := range importPaths {
		err := afero.Walk(fsys, importPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("failed to walk %q path: %w", importPath, err)
			}

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
				protoFiles = append(protoFiles, strings.TrimPrefix(path, importPath))
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk import path %q: %w", importPath, err)
		}
	}

	return protoFiles, nil
}
