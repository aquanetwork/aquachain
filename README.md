# Aquachain

## General Purpose Smart Contract Distributed Super Computer

Mining Reward: Only one AQUA reward every block.

Block time: 4 minutes

Supply Cap: At block 42mil block rewards become zero (gas only)

## GET AQUACHAIN


### Install

To begin, you must have aquachain command installed.

To build with go, use `go get -v github.com/aquanetwork/aquachain/cmd/aquachain`

To download binary release, see [Releases](https://github.com/aquanetwork/aquachain/releases/)

Apple/Linux Install: (Apple users, replace linux-amd64 with darwin-amd64)

	sudo chmod +x aquachain-linux-amd64 			# make executable
	sudo mv aquachain-linux-amd64 /usr/local/bin/aquachain 	# install
	aquachain account new					# create new private/public key pair
	aquachain --mine console				# start mining, open console

Windows:
	unzip where you want. 
	right click the exe and make a shortcut.
	right click the shortcut and add the word `console`
	Should look like `c:\path\to\aquachain.exe" console`

### Update

the aquachain command is a portable program that can be moved anywhere.

delete the .exe and replace with new version. the exe is safe to delete.

your wallet and blockchain exist in a hidden folder, dont delete that folder!

the best way to stay up to date is build your self, using [golang](https://golang.org/doc/install)

### Wallet/Blockchain is stored in `datadir`

If you need to back up your wallet, or use the json file in MEW, the location is:

Linux, Mac: `~/.aquachain/`

Windows: type `%appdata%\AquaChain` in the windows explorer


	
## MINE AQUACHAIN

ONE TWO THREE

NOTE: If you see any errors, see TROUBLESHOOTING section. Instructions are for linux and/or apple. Windows users install to C:\Program Files\aquachain.exe

NOTE: If you are receiving failed messages when connecting to the network please download the bootstrap.dat file (you can find at the bottom of the readme) and add it to the same directory as the executable. From the command prompt type: 'aquachain import bootstrap.dat'

1)Create an account: This will create a wallet address where the mining rewards are stored. By default the last address created is where the mining rewards are sent. Type the following:

```
aquachain account new
```

2)After creating a password for the wallet, start the AquaChain console by running the following:

```
aquachain console
```

3)While in the console, there are many commands you can run. Before mining, it is recommended to check how many peers you are connected to. In order to do this, type admin.peers.length after connecting to the console. Any number greater than 0 shows you have access to the network. To begin mining type the following in the console:

```
> miner.start()
```

This will start by generating the DAG, which can take up to 30 minutes or more.

### cpu #

To specify the number of CPUs to use while mining, add a number in between the parenthesis. For example to run only 1 cpus, type miner.start(1). Dual core has 2 cpu. Quad core has 4 etc.

### Use JSON-RPC for local copy of block explorer, or MEW.

If you want to use the json-rpc, you must use a command line such as this:

```
aquachain --rpc --rpccorsdomain='*' --rpcvhosts='*' --rpcapi 'eth,aqua,web3' console
```

This will listen on port 8543, so make sure to use 'http://localhost' and '8543' when setting up MEW custom node.

## Resources

Explorer - http://explorer.aquanetwork.co/
Download Latest: https://github.com/aquanetwork/aquachain/releases
More Downloads - http://explorer.aquanetwork.co/dl/
Github - http://github.com/aquachain
Github - http://github.com/aquanetwork/aquachain
News, Chat: https://t.me/AquaCrypto

Contributions welcome. Check out @AquaCrypto on telegram for ways to help.

[![Build Status](https://travis-ci.org/aquanetwork/aquachain.svg?branch=master)](https://travis-ci.org/aquanetwork/aquachain)

