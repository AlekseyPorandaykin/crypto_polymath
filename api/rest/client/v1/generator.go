package v1

//go:generate go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v2.4.0

//go:generate oapi-codegen -old-config-style -package v1 -generate models,skip-prune -o models.go  ../../rest/v1/openapi.yaml
