.PHONY: protoc
protoc:
	protoc --go_out=internal/testdata --go-grpc_out=internal/testdata -I=internal/testdata --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative internal/testdata/test.proto
