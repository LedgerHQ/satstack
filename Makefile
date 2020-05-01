dev:
	go run cmd/main.go

build:
	go build -v cmd/main.go

regtest:
	go test -v -timeout 0 ./tests/regression/...
