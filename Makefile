# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: aquachain android ios aquachain-cross swarm evm all test clean
.PHONY: aquachain-linux aquachain-linux-386 aquachain-linux-amd64 aquachain-linux-mips64 aquachain-linux-mips64le
.PHONY: aquachain-linux-arm aquachain-linux-arm-5 aquachain-linux-arm-6 aquachain-linux-arm-7 aquachain-linux-arm64
.PHONY: aquachain-darwin aquachain-darwin-386 aquachain-darwin-amd64
.PHONY: aquachain-windows aquachain-windows-386 aquachain-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

aquachain:
	build/env.sh go run build/ci.go install ./cmd/aquachain
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquachain\" to launch aquachain."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

release: aquachain-windows-amd64 aquachain-darwin-amd64 aquachain-linux-amd64


android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/aquachain.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/AquaChain.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

aquachain-cross: aquachain-linux aquachain-darwin aquachain-windows aquachain-android aquachain-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-*

aquachain-linux: aquachain-linux-386 aquachain-linux-amd64 aquachain-linux-arm aquachain-linux-mips64 aquachain-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-*

aquachain-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/aquachain
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep 386

aquachain-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/aquachain
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep amd64

aquachain-linux-arm: aquachain-linux-arm-5 aquachain-linux-arm-6 aquachain-linux-arm-7 aquachain-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep arm

aquachain-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/aquachain
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep arm-5

aquachain-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/aquachain
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep arm-6

aquachain-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/aquachain
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep arm-7

aquachain-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/aquachain
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep arm64

aquachain-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/aquachain
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep mips

aquachain-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/aquachain
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep mipsle

aquachain-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/aquachain
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep mips64

aquachain-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/aquachain
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-linux-* | grep mips64le

aquachain-darwin: aquachain-darwin-386 aquachain-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-darwin-*

aquachain-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/aquachain
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-darwin-* | grep 386

aquachain-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/aquachain
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-darwin-* | grep amd64

aquachain-windows: aquachain-windows-386 aquachain-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-windows-*

aquachain-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/aquachain
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-windows-* | grep 386

aquachain-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/aquachain
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquachain-windows-* | grep amd64
