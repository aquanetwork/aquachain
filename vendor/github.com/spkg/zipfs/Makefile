.PHONY: \
	build \
	test \
	cover

build:
	go build

test:
	go test

cover:
	go test -coverprofile coverage.out
	go tool cover -html coverage.out
