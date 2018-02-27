# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: aquad android ios aquad-cross swarm evm all test clean
.PHONY: aquad-linux aquad-linux-386 aquad-linux-amd64 aquad-linux-mips64 aquad-linux-mips64le
.PHONY: aquad-linux-arm aquad-linux-arm-5 aquad-linux-arm-6 aquad-linux-arm-7 aquad-linux-arm64
.PHONY: aquad-darwin aquad-darwin-386 aquad-darwin-amd64
.PHONY: aquad-windows aquad-windows-386 aquad-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

aquad:
	build/env.sh go run build/ci.go install ./cmd/aquad
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquad\" to launch aquad."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/aquad.aar\" to use the library."

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

aquad-cross: aquad-linux aquad-darwin aquad-windows aquad-android aquad-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/aquad-*

aquad-linux: aquad-linux-386 aquad-linux-amd64 aquad-linux-arm aquad-linux-mips64 aquad-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-*

aquad-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/aquad
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep 386

aquad-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/aquad
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep amd64

aquad-linux-arm: aquad-linux-arm-5 aquad-linux-arm-6 aquad-linux-arm-7 aquad-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep arm

aquad-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/aquad
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep arm-5

aquad-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/aquad
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep arm-6

aquad-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/aquad
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep arm-7

aquad-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/aquad
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep arm64

aquad-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/aquad
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep mips

aquad-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/aquad
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep mipsle

aquad-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/aquad
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep mips64

aquad-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/aquad
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/aquad-linux-* | grep mips64le

aquad-darwin: aquad-darwin-386 aquad-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/aquad-darwin-*

aquad-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/aquad
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-darwin-* | grep 386

aquad-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/aquad
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-darwin-* | grep amd64

aquad-windows: aquad-windows-386 aquad-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/aquad-windows-*

aquad-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/aquad
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-windows-* | grep 386

aquad-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/aquad
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/aquad-windows-* | grep amd64
