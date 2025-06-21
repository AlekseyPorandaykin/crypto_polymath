package spec

//go:generate go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v2.4.0

//go:generate oapi-codegen -old-config-style -package spec -generate types,skip-prune -o types.gen.go  ../../../../../api/rest/v1/openapi.yaml
//go:generate oapi-codegen -old-config-style -package spec -generate server  -o server.gen.go ../../../../../api/rest/v1/openapi.yaml
