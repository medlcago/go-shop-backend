doc:
	swag fmt && swag init -g cmd/api/main.go --dir ./ --parseDependency --parseInternal --parseDepth 5 --output ./docs

test:
	go test ./...

mocks:
	mockery --config .mockery.yaml
