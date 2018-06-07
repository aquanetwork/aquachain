.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    gitlab.com/aquachain/aquachain/cmd/swarm
	go build -o bin/aquachain     gitlab.com/aquachain/aquachain/cmd/aquachain
	go build -o bin/bootnode gitlab.com/aquachain/aquachain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v gitlab.com/aquachain/aquachain/opt/swarm/...
