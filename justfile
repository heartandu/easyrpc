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
    @protoc --go_out=internal/testdata --go-grpc_out=internal/testdata -I=internal/testdata --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative internal/testdata/test.proto
