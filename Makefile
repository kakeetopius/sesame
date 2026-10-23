vet:
	go vet ./...
start-local: clean-db 
	env=local go run main.go --config samples/sesame.toml
start-server:
	go run cmd/main.go --config samples/sesame.toml
build-linux:
	go build -o build/sesame .
local-build:
	go build -o build/sesame .
clean-db:
	rm -rf db
test:
	env=local go test -v -failfast -race ./...
