# Aquachain

Latest Source: https://gitlab.com/aquachain/aquachain

How to build: https://github.com/aquanetwork/aquachain/wiki/Compiling

See bottom of this document for more useful links.

## General Purpose Distributed Computing

ETH compatible smart contract platform

Target Block Time: 240 second blocks (4 minute)

Block Reward: 1 AQUA

Max Supply: 42 million

Explorer: https://aquachain.github.io/explorer/

Algorithm: argon2id (CPU minable)

Use MyEtherWallet or Metamask to connect to an SSL aquachain node like `https://c.onical.org`

(EIP 155) Chain ID: 61717561

## GET AQUACHAIN

To begin, you must have the aquachain command installed. The `aquachain` command is a portable program that doesn't really need an 'installer', you can run it from anywhere.

To build latest with go (recommended), use 'go get':

`go get -v -u gitlab.com/aquachain/aquachain/cmd/aquachain` and for the miner `go get -v -u gitlab.com/aquachain/aquachain/cmd/aquaminer`

To download binary release, see [Releases](https://github.com/aquanetwork/aquachain/releases/latest) tab on github.

Create account inside the console: `personal.newAccount()`

Your new wallet private key is located inside `datadir` by default is `~/.aquachain/keystore` or `%appdata%\Roaming\AquaChain` (windows), make backup(s) and don't forget the pass phrase!

See more commands: [Wiki](https://github.com/aquanetwork/aquachain/wiki/Basics)

### New Default Console Mode

Now double-clickable! Just unzip and run to enter the Aquachain Javascript Console

Type `help` at the `AQUA>` prompt for common commands.

For automated scripts and whatnot, add 'daemon' argument for the previous default action:

```
aquachain -rpc -rpcapi 'aqua,eth,net,web3' daemon
```

## Resources

Wiki - https://github.com/aquanetwork/aquachain/wiki

Website - https://aquachain.github.io

ANN - https://bitcointalk.org/index.php?topic=3138231.0

Explorer - https://aquachain.github.io/explorer/

Github - http://github.com/aquachain

Gitlab - http://gitlab.com/aquachain/aquachain

Telegram Chat: https://t.me/AquaCrypto

Telegram News: https://t.me/Aquachain

Godoc - https://godoc.org/gitlab.com/aquachain/aquachain#pkg-subdirectories

Report bugs - https://github.com/aquachain/aquachain/issues

Discord: https://discordapp.com/invite/J7jBhZf

## Contributing

Contributions welcome. Check out @AquaCrypto on telegram for ways to help.

[![Build Status](https://travis-ci.org/aquanetwork/aquachain.svg?branch=master)](https://travis-ci.org/aquanetwork/aquachain)
