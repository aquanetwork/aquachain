# Aquachain

Latest Source: https://gitlab.com/aquachain/aquachain 
How to build: https://github.com/aquanetwork/aquachain/wiki/Compiling

See bottom of this document for more useful links. Contributions are welcome.

## General Purpose Distributed Computing

Aquachain: peer-to-peer programmable money, distributed code contract platform.

  Target Block Time: 240 second blocks (4 minute)
  Block Reward: 1 AQUA
  Max Supply: 42 million 
  Explorer: https://aquachain.github.io/explorer/ 
  Algorithm: argon2id (CPU mined)
  ChainID/NetworkID: 61717561

## GET AQUACHAIN

To begin, you must have the aquachain command. The `aquachain` command is a
portable program that doesn't really need an 'installer', you can run it from
anywhere. When you first start `aquachain` you will connect to the peer-to-peer
network and start downloading the chain. To change the way aquachain runs, for
example testnet, or rpc, use command line flags. List all command line flags
using the `-h` flag, or `aquachain help [subcommand]`

## COMPILING

If you are reading this from the source tree, you can `go build ./cmd/aquachain`

** Bugs can be reported at https://github.com/aquanetwork/aquachain/issues/ **

To [build latest](Documentation/Compiling.md) with go (recommended), simply use
'go get' and look in $GOPATH/bin:

	CGO_ENABLED=0 go get -v -u gitlab.com/aquachain/aquachain/cmd/aquachain

or all tools:

	GOBIN=$PWD go get -v gitlab.com/aquachain/aquachain/cmd/...

To see latest release, check `git log` or:

    [Releases](https://github.com/aquanetwork/aquachain/releases/latest)

## SYNCHRONIZING

"Imported new chain segment" means you received new blocks from the network.
When a single block is imported, the address of the successful miner is printed.
When you start seeing one every 4 minutes or so, you are fully synchronized.

## USAGE

Create account from the command line: `aquachain.exe account new`
List accounts from the command line: `aquachain.exe account list`
Enter AQUA console: `aquachain.exe`
Start Daemon (geth default): `aquachain.exe daemon`
See more commands: [Wiki](https://github.com/aquanetwork/aquachain/wiki/Basics)
Type `help` at the `AQUA>` prompt for common AQUA console commands.

Start HTTP JSON/RPC server: `aquachain -rpc`

Note about RPC: By default, the `-rpc` flag Please be aware that hosting a
public RPC server (0.0.0.0) will allow strangers access to your system. Do not
use the `-rpcaddr`flag unless you absolutely know what you are doing.

For hosting public RPC servers, consider using -nokeys and implementing 
rate limiting (via reverse proxy such as caddyserver or nginx).

The JSON/RPC server is able to be used with "Web3" libraries for languages such
as Python or Javascript. Go packages for Aquachain RPC client can be found in
this repository, under the `opt/aquaclient` namespace. See package documentation
for more information on usage.

## Resources

Wiki - https://github.com/aquanetwork/aquachain/wiki
Website - https://aquachain.github.io
ANN - https://bitcointalk.org/index.php?topic=3138231.0
Explorer - https://aquachain.github.io/explorer/
Github - http://github.com/aquachain
Gitlab - http://gitlab.com/aquachain/aquachain
Telegram News: https://t.me/Aquachain
Godoc - https://godoc.org/gitlab.com/aquachain/aquachain#pkg-subdirectories
Report bugs - https://github.com/aquachain/aquachain/issues
Telegram Chat: https://t.me/AquaCrypto
Discord: https://discordapp.com/invite/J7jBhZf
IRC: #aquachain on freenode

## Contributing

Aquachain is free open source software and your contributions are welcome.


[![Build Status](https://travis-ci.org/aquanetwork/aquachain.svg?branch=master)](https://travis-ci.org/aquanetwork/aquachain)

### Some tips and tricks for hacking on Aquachain core:

  * Always `gofmt -w -l -s` before commiting. If you forget, adding a simple
    'gofmt -w -l -s' commit works.
  * `AQUAPATH=$(go env GOPATH)/src/gitlab.com/aquachain/aquachain` in
    ~/.bashrc, this saves time.  Work in $AQUAPATH, and use `git branch` to
    navigate git forks (`git remote add fork
    git@github.com:user/aquachain.git`), this prevents having to change import
    paths.
  * Before making a pull request, try `make test` to run all tests. If any
    tests pass, the PR can not be merged into the master branch.
  * Rebase: Don't `git pull`, use `git pull -r` or `git rebase -i master` from
    your branch
  * Squash same-file similar commits if possible
  * Prefix commit message with package name, such as "core: fix blockchain"
