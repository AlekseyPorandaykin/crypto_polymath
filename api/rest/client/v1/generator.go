package v1

//go:generate go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.15.0

//go:generate oapi-codegen -old-config-style -package v1 -generate models,skip-prune -o models.go  ../../v1/openapi.yaml
