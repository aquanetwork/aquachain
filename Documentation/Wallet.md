#### **First Boot**

To get started, build (see [Compiling](Compiling)) or download the latest release **only from** https://github.com/aquanetwork/aquachain/releases

Unzip the archive, and double click the `aquachain.exe` application to open the javascript console.

When you open the program, you immediately start finding nodes and synchronizing with the network. It should take under 5 minutes to sync the entire chain at this time. 


## Startup Options

Some times you might need to start aquachain with a certain command, or flags:


### Example command arguments
```
aquachain.exe account new # create new account
aquachain.exe -h # show help
aquachain.exe version # show version
aquachain.exe daemon # no console
aquachain.exe removedb # delete blockchain (keeping private keys)
aquachain.exe wallet # opens browser based wallet
aquachain.exe paper 10 ## generates ten addresses
aquachain.exe paper -vanity 123 10 ## generates ten addresses beginning with 0x123
```

### Disabling P2P
To disable p2p, use `-nodiscover` and `-maxpeers 0` flags


## Aquachain Console

You know if you are in the aquachain console if you see the **AQUA>** prompt. It is a javascript console, with (tab) auto-complete and (up/down) command history.


  * Load a local script with the `loadScript('filename')` function.
  * List accounts with `aqua.accounts`
  * Check all balances with `balance()`



### **Wallet**

To store your aqua, you need an account. 

One can be generated in three ways:

* `aquachain.exe account new`
* [aquapaper.exe](https://github.com/aquanetwork/aquachain/releases/tag/paper-0.0) paper wallet generator (now built-in! use `aquachain.exe paper`)
* inside the experimental [webwallet](https://github.com/aquanetwork/aquachain/releases/tag/paper-0.0) browser app (now built-in! use `aquachain.exe wallet`) and connect to your local instance

### Send a transaction

Start a transaction by typing `send` and press enter (type `n` or `ctrl+c` to cancel)


### Most important commands

Check balance with `aqua.balance(aqua.coinbase)`

Check all accounts balances: `balance()`

Send a transaction (easy way): `send`

### Sending a transaction the hard way

Before sending coins, you must unlock the account:

`personal.unlockAccount(aqua.accounts[0])`

This command will send 2 AQUA from "account 0" (first account created) to the typed account below:
```
aqua.sendTransaction({from: aqua.accounts[0], to: '0x3317e8405e75551ec7eeeb3508650e7b349665ff', value:web3.toWei(2,"aqua")});
```

Since its a javascript console, you can do something like this:

```
var destination = '0x3317e8405e75551ec7eeeb3508650e7b349665ff';
aqua.sendTransaction({from: aqua.accounts[0], to: destination, value:web3.toWei(2,"aqua")});
```

Default gas price is 0.1 gwei, if you want to specify a higher gas price you can add it to the tx:

```
var tx = {from: aqua.accounts[0], to: destination, gasPrice: web3.toWei(20, "gwei"), value:web3.toWei(2,"aqua")};
aqua.sendTransaction(tx);
```

### Useful Console Commands 

(Save time! Press tab twice to auto-complete everything.)

```
balance()
aqua.balance(aqua.coinbase)
send
admin.nodeInfo.enode
net.listening
net.peerCount
admin.peers
aqua.coinbase
aqua.getBalance(aqua.coinbase)
personal
aqua.accounts
miner.setAquabase(web3.aqua.accounts[0])
miner.setAquabase(“0x0000000000000000000000000000000000000000”)
miner.start()
miner.stop()
miner.hashrate
aqua.getBlock(0)
aqua.getBlock(“latest”)
aqua.blockNumber 
web3.aqua.getBlock(BLOCK_NUMBER).hash
aqua.syncing
debug.verbosity(6) // highest logging level, 3 is default
```


### Import Private Key

If you created a wallet with `aquachain.exe paper` or `aquachain wallet` you will have an unprotected private key. To send coins from the account derived from this key, you must import it into the console.

to import a private key into the aquachain console:

prepare your private key as a **simple text file with no spaces**. name it anything.txt for example

run `aquachain.exe account import anything.txt`

make sure to delete/shred the private key if it has successfully been imported.
