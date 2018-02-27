.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/aquanetwork/aquachain/cmd/swarm
	go build -o bin/aquad     github.com/aquanetwork/aquachain/cmd/aquad
	go build -o bin/bootnode github.com/aquanetwork/aquachain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/aquanetwork/aquachain/swarm/...
