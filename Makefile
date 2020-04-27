dev:
	go run cmd/main.go

build:
	go build -v cmd/main.go

regtest:
	go test tests/regression_test.go -v
