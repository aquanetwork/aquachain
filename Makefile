GOBIN = ${PWD}/build/bin
GO ?= latest
PREFIX ?= ${HOME}/.local/bin/
CGO_ENABLED ?= "0"

# make build environment script executable (gets unset through ipfs)
DOFIRST != chmod +x build/env.sh

# default build
COMMITHASH != git rev-parse HEAD
aquachain:
	@echo "Building default aquachain: ./build/bin/aquachain"
	GOBIN=${GOBIN} CGO_ENABLED=${CGO_ENABLED} go install -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

cross: aquachain-win32.exe aquachain-win64.exe aquachain-linux-386 aquachain-linux-amd64 aquachain-arm aquachain-osx aquachain-freebsd aquachain-openbsd aquachain-netbsd

aquachain-win32.exe:
	GOOS=windows GOARCH=386 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-win32.exe -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-win64.exe:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-win64.exe -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-linux-386:
	GOOS=linux GOARCH=386 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-linux-386 -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-linux-amd64 -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-arm:
	GOOS=linux GOARCH=arm CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-arm -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-osx:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-osx -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-freebsd:
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-freebsd -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-openbsd:
	GOOS=openbsd GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-openbsd -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain

aquachain-netbsd:
	GOOS=netbsd GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -o aquachain-netbsd -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain



nocache:
	GOBIN=${GOBIN} CGO_ENABLED=${CGO_ENABLED} go install -a -tags 'netgo osusergo' -ldflags '-X main.gitCommit=${COMMITHASH} -s -w' -v ./cmd/aquachain


# this is old way
aquachain-go:
	@echo "Building aquachain with no tracer/usb support."
	@echo "Consider \"${MAKE} usb\" or \"${MAKE} aquachain\""
	@echo "Building default aquachain. Consider \"${MAKE} musl\""
	CGO_ENABLED=${CGO_ENABLED} build/env.sh go run build/ci.go install ./cmd/aquachain
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquachain\" to launch aquachain."

aquachain-cgo:
	@echo "Building aquachain with no usb support. Consider \"${MAKE} usb\""
	@echo "Building default aquachain. Consider \"${MAKE} musl\""
	CGO_ENABLED=1 build/env.sh go run build/ci.go install ./cmd/aquachain
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquachain\" to launch aquachain."

# with usb support (hardware wallet)
usb:
	build/env.sh go run build/ci.go install -usb ./cmd/aquachain 

# static, using musl c lib
musl:
	build/env.sh go run build/ci.go install -static -musl ./cmd/aquachain 

# static linked binary
static:
	build/env.sh go run build/ci.go install -static ./cmd/aquachain

# build (WIP) reference stratum client
aquastrat:
	@echo "Building aquastrat, stratum test client"
	build/env.sh go run build/ci.go install -static ./cmd/aquastrat

# build reference miner
aquaminer:
	CGO_ENABLED=${CGO_ENABLED} build/env.sh go run build/ci.go install ./cmd/aquaminer
	@echo "Done building."
	@echo "Run \"$(GOBIN)/aquaminer\" to start mining to localhost:8543 rpc-server."

# build all tools also see aquachain/x repo
all:
	build/env.sh go run build/ci.go install

# all tools linked statically
all-static:
	build/env.sh go run build/ci.go install -static

# all tools built with musl
all-musl:
	build/env.sh go run build/ci.go install -musl -static

# ci/test stuff

test: all
	build/env.sh go run build/ci.go test

test-verbose: all
	build/env.sh go run build/ci.go test -v

test-race: all
	build/env.sh go run build/ci.go test -race

test-musl: musl
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
	install $(GOBIN)/* $(PREFIX)

.PHONY: aquachain all test clean
.PHONY: aquaminer aquastrat race install generate lint musl all-musl static
.PHONY: docker-run cross generate devtools

docker-run:
	mkdir -p ${HOME}/.aquachain-alt
	docker run -it -p 127.0.0.1:8543:8543 -v ${HOME}/.aquachain-alt/:/root/.aquachain aquachain/aquachain:latest -- aquachain -rpc

#cross:
#	xgo -image aquachain/xgo -ldflags='-w -s -extldflags -static' -tags 'osusergo netgo static' -pkg cmd/aquachain -targets='windows/*,linux/arm,linux/386,linux/amd64,darwin/amd64' gitlab.com/aquachain/aquachain

# this builds test-binaries to remove compilation time between repeating tests
debugging:
	CGO_ENABLED=1 go test -race -o tester-race ./...
	CGO_ENABLED=0 go test -o tester-nocgo ./...
	CGO_ENABLED=1 go test -o tester-cgo ./...
