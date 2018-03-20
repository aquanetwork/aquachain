# Aquachain

## General Purpose Distributed Computing

Block time: 240 second blocks (4 minute)

Block reward: Ⱥ1.00

Max supply: 42 million

Deploy and interact with smart contracts

## GET AQUACHAIN

To begin, you must have the aquachain command installed.

To build latest with go, use `go get -v github.com/aquanetwork/aquachain/cmd/aquachain`

To download binary release, see [Releases](https://github.com/aquanetwork/aquachain/releases/latest)

The `aquachain` command is a portable program that doesn't really need an 'installer', you can run it from anywhere.

Apple/Linux Install: (Apple users, replace linux-amd64 with darwin-amd64)

	sudo mv aquachain /usr/local/bin/aquachain 	# install
	aquachain account new					# create new private/public key pair
	aquachain console				      # open console

Create an account: This will create a wallet address where the mining rewards are stored. By default the last address created is where the mining rewards are sent. Type the following:

```
aquachain account new
```

Your new wallet private key is located inside `datadir` by default is `~/.aquachain/`` or ``%appdata%\Roaming\AquaChain` (windows)

After creating a password for the wallet, start the AquaChain console by running the following:

```
aquachain console
```

Type `help` at the `AQUA>` prompt for common commands.

## Resources

Wiki - https://github.com/aquanetwork/aquachain/wiki
Explorer - http://explorer.aquanetwork.co/
Downloads - http://explorer.aquanetwork.co/dl/
Github - http://github.com/aquachain
Github - http://github.com/aquanetwork/aquachain
News, Chat: https://t.me/AquaCrypto
Godoc - http://godoc.org/github.com/aquanetwork/aquachain
Report bugs - https://github.com/aquanetwork/aquachain/issues
Discord: https://discordapp.com/invite/J7jBhZf

Contributions welcome. Check out @AquaCrypto on telegram for ways to help.

[![Build Status](https://travis-ci.org/aquanetwork/aquachain.svg?branch=master)](https://travis-ci.org/aquanetwork/aquachain)
