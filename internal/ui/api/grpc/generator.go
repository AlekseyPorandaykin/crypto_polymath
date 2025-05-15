package grpc

//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//go:generate protoc --go_out=./internal/ui/api/grpc/action  --go-grpc_out=./internal/ui/api/grpc/action   ./api/grpc/v1/ActionService.proto
