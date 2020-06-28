dev:
	go run cmd/lss.go

release:
	GIN_MODE=release go run cmd/lss.go

build:
	go build -v cmd/lss.go

regtest:
	go test -v -timeout 0 ./tests/regression/...

it:
	go test -v -timeout 0 ./tests/integration/...
