GOBIN = $(shell pwd)/build/bin
GO ?= latest
PREFIX ?= ${HOME}/.local/bin/

# default build
aquachain:
	@echo "Building aquachain with no usb support. Consider \"${MAKE} usb\""
	@echo "Building default aquachain. Consider \"${MAKE} musl\""
	build/env.sh go run build/ci.go install ./cmd/aquachain
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquachain\" to launch aquachain."


# with usb support (hardware wallet)
usb:
	build/env.sh go run build/ci.go install -usb ./cmd/aquachain 

# static linked binary
static:
	build/env.sh go run build/ci.go install -static ./cmd/aquachain

# build reference miner
aquaminer:
	build/env.sh go run build/ci.go install ./cmd/aquaminer
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquaminer\" to start mining to localhost:8543 rpc-server."

# experimental
browsermine:
	GOARCH=wasm GOOS=js go build -o aquaminer.wasm ./cmd/aquaminer
	@echo "Done building."

# build all tools also see aquachain/x repo
all:
	build/env.sh go run build/ci.go install

# all tools linked statically
all-static:
	build/env.sh go run build/ci.go install -static

# all tools built with musl
all-musl:
	build/env.sh go run build/ci.go install -musl -static

# unused

release: aquachain-windows-amd64 aquachain-darwin-amd64 aquachain-linux-amd64

archive:
	build/env.sh go run build/ci.go archive

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/aquachain.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/AquaChain.framework\" to use the library."

# ci/test stuff

test: all
	build/env.sh go run build/ci.go test 

musl-test: musl
	build/env.sh go run build/ci.go test -musl 

lint: 
	build/env.sh go run build/ci.go lint
clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*
	rm -fr build/_workspace/src/ $(GOBIN)/*
	rm -fr /tmp/aqua/_workspace/pkg/ $(GOBIN)/*
	rm -fr /tmp/aqua/_workspace/src/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get golang.org/x/tools/cmd/stringer
	env GOBIN= go get github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get github.com/fjl/gencodec
	env GOBIN= go get github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install gitlab.com/aquachain/x/cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

generate: devtools
	go generate ./...
# build binary that can detect race conditions
race:
	build/env.sh go run build/ci.go install -- -race ./cmd/aquachain/

# install to $(PREFIX)
install:
	install $(GOBIN)/aquachain $(PREFIX)

.PHONY: aquachain android ios aquachain-cross swarm evm all test clean
.PHONY: aquachain-linux aquachain-linux-386 aquachain-linux-amd64 aquachain-linux-mips64 aquachain-linux-mips64le
.PHONY: aquachain-linux-arm aquachain-linux-arm-5 aquachain-linux-arm-6 aquachain-linux-arm-7 aquachain-linux-arm64
.PHONY: aquachain-darwin aquachain-darwin-386 aquachain-darwin-amd64
.PHONY: aquachain-windows aquachain-windows-386 aquachain-windows-amd64
.PHONY: aquaminer
