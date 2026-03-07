default:
    @just --list

run *args:
    go run ./cmd/easyrpc {{args}}

build:
    go build -o bin/easyrpc ./cmd/easyrpc

test *flags:
    go test -count=1 -race -coverprofile coverage.out {{flags}} ./...

lint:
    golangci-lint run

protoc:
    @protoc --go_out=internal/testdata/proto --go-grpc_out=internal/testdata/proto -I=internal/testdata/proto \
    --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
    internal/testdata/proto/types/types.proto internal/testdata/proto/types/common/common.proto \
    internal/testdata/proto/echo/echo.proto internal/testdata/proto/packageless.proto

gen-certs:
    cd internal/testdata && zsh ../../gen_certs.sh
