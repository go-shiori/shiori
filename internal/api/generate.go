package api

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=api --generate=types -o types.gen.go ../../openapi.yml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=api --generate=server,spec -o spec.gen.go ../../openapi.yml
