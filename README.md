# Aquachain

## General Purpose Distributed Computing

Target Block Time: 240 second blocks (4 minute)

Block Reward: Ⱥ1.00

Max Supply: 42 million

HF5: 22800

## GET AQUACHAIN

To begin, you must have the aquachain command installed. The `aquachain` command is a portable program that doesn't really need an 'installer', you can run it from anywhere.

To build latest with go, use `go get -v github.com/aquanetwork/aquachain/cmd/aquachain`

To download binary release, see [Releases](https://github.com/aquanetwork/aquachain/releases/latest)

Apple/Linux Install: (Apple users, replace linux-amd64 with darwin-amd64)

	sudo mv aquachain /usr/local/bin/aquachain  # install
	aquachain account new			    # create new private/public key pair
	aquachain            			    # open console

Create account inside the console: `personal.newAccount()`

Your new wallet private key is located inside `datadir` by default is `~/.aquachain/`` or ``%appdata%\Roaming\AquaChain` (windows)

See more commands: [Wiki](https://github.com/aquanetwork/aquachain/wiki/Basics)

### New Default Console Mode

Now double-clickable! Just unzip and run to enter the Aquachain Javascript Console

Type `help` at the `AQUA>` prompt for common commands.

### Init scripts / No Console

For automated scripts and whatnot, add 'daemon' argument for the previous default action:

```
aquachain daemon
```

## Resources

Wiki - https://github.com/aquanetwork/aquachain/wiki

Explorer - http://explorer.aquanetwork.co/

Downloads - http://explorer.aquanetwork.co/dl/

Github - http://github.com/aquachain

Github - http://github.com/aquanetwork/aquachain

Telegram Chat: https://t.me/AquaCrypto

Telegram News: https://t.me/Aquachain

Godoc - https://godoc.org/github.com/aquanetwork/aquachain#pkg-subdirectories

Report bugs - https://github.com/aquanetwork/aquachain/issues

Discord: https://discordapp.com/invite/J7jBhZf

## Contributing

Contributions welcome. Check out @AquaCrypto on telegram for ways to help.

[![Build Status](https://travis-ci.org/aquanetwork/aquachain.svg?branch=master)](https://travis-ci.org/aquanetwork/aquachain)
