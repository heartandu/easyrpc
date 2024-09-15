package proto

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"google.golang.org/grpc"

	"github.com/heartandu/easyrpc/internal/config"
	"github.com/heartandu/easyrpc/pkg/descriptor"
)

// NewDescriptorSource returns a new descriptor source based on the provided configuration.
func NewDescriptorSource(
	ctx context.Context,
	fs afero.Fs,
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
		descSrc, err = descriptor.ProtoFilesSource(ctx, fs, cfg.Proto.ImportPaths, cfg.Proto.ProtoFiles)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create descriptor source: %w", err)
	}

	return descSrc, nil
}
