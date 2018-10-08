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

The `aquachain` command is a portable program that doesn't really need an
'installer', you can run it from anywhere. When you first start `aquachain` you
will connect to the peer-to-peer network and start downloading the chain. To
change the way aquachain runs, for example testnet, or rpc, use command line
flags. The location of your keys can be printed with: `aquachain account list`

List all command line flags using the `-h` flag, or `aquachain help [subcommand]`

## COMPILING

If you are reading this from the source tree, you can `go build ./cmd/aquachain`

** Bugs can be reported at https://github.com/aquanetwork/aquachain/issues/ **

** Patches can be submitted at Github or Gitlab or Mailing List **

To [build latest](Documentation/Compiling.md) with go (recommended), simply use
'go get' and look in $GOPATH/bin:

	CGO_ENABLED=0 go get -v -u gitlab.com/aquachain/aquachain/cmd/aquachain

or all tools:

	GOBIN=$PWD go get -v gitlab.com/aquachain/aquachain/cmd/...

To see latest release, check `git log` or:

  * [Releases](https://github.com/aquanetwork/aquachain/releases/latest)

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


## RPC SERVER

See "RPC" section in ./Documentation folder and online at:
https://github.com/aquanetwork/aquachain/wiki/RPC

Start HTTP JSON/RPC server for local (127.0.0.1) connections only:
	
	aquachain -rpc

Start HTTP JSON/RPC server for remote connections, listening on 192.168.1.5:8543,
able to be accessed only by 192.168.1.6:

	aquachain -rpc -rpchost 192.168.1.5 -allowip 192.168.1.6

Security Note about RPC: By default, the `-rpc` flag Please be aware that hosting a
public RPC server (0.0.0.0) will allow strangers access to your system. Do not
use the `-rpcaddr`flag unless you absolutely know what you are doing.

For hosting public RPC servers, please consider using -nokeys (*new!*) and implementing
rate limiting on http (and, if using, websockets) , either via reverse proxy such as
caddyserver or nginx, or firewall.

Recent builds of aquachain include support for the `-allowip` flag. It is by default,
set to 127.0.0.1, which doesn't allow any LAN or WAN addresses access to your RPC methods.

To add IPs, use  `aquachain -rpc -rpchost 192.168.1.4 -allowip 192.168.1.5,192.168.1.6`

Support for CIDR notation would be great, but did not make it into the current 
implementation. The IP addresses are comma separated, no spaces.

#### RPC Clients

The JSON/RPC server is able to be used with "Web3" libraries for languages such
as **Python** or **Javascript**. 

All calls to `eth_` methods are translated to `aqua_`, server-side.

**Go** packages for creating applications that use Aquachain can be found in
this repository, under the `opt/aquaclient` and `rpc/rpcclient` namespaces. 
See each package's documentation (godoc) for more information on usage.

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
