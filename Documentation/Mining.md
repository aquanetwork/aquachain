## New Mining Software

ATTN: do not use aquaminer unless testing. you wont get good hashrates compared to the c++ miner.

now, use aquacppminer.

Grab it here: https://bitbucket.org/cryptogone/aquacppminer/downloads

## Aquachain Miner Resources

  * [Hashrate Distribution](https://explorer.aqua.signal2noi.se/stats/miner_hashrate)
  * [Network Pool Status](https://aquacha.in/status/miners)
  * [Mining Calculator](https://docs.google.com/spreadsheets/d/1MIe8YDY8ORBDukZDmlrG6QQx1Fw0wckKZxh4pJUDAxw/edit?usp=sharing)

[Pool Mining](#pool-mining)

[Solo Mining](#solo-mining)

## Pool mining

### Create new wallet

Note: linux and osx (darwin) users can just remove the '.exe' from instructions.

Grab [aquachain.exe](https://github.com/aquanetwork/aquachain/releases/)

Unzip, and optionally rename from aquachain-1.7.0-windows-amd64.exe to aquachain.exe

`aquachain.exe account new` (console wallet)

or 

`aquachain.exe paper` (not password encrypted)

Make sure to set a good password. Don't forget it. Be your own bank.

When using `aquachain.exe account new`, (or from within the AQUA console: `personal.newAccount()`), your wallet private key is encrypted with a pass phrase and saved in your "keystore" directory. 

Your public address is printed after creating an account. If using `account new` you will see something like:

```
Your new account is locked with a password. Please give a password. Do not forget this password.
Passphrase: 
Repeat passphrase: 
Address: {1a7a0c0fd8d138f132b6a2ce22a715abebc16742}
```

Add `0x` and that is your address. Like this: `0x1a7a0c0fd8d138f132b6a2ce22a715abebc16742`

Once you are mining, see [Wallet](Wallet) for more information about using the wallet.

Download the [aquacppminer](https://bitbucket.org/cryptogone/aquacppminer/downloads) tool.

### Command for mining

Choose a pool from those listed on [pool status](https://aquacha.in/status/miners) or [aquachain.github.io](https://aquachain.github.io) or check our bitcoin talk [ANN](https://bitcointalk.org/index.php?topic=3138231.new#new) for newer ones.

Replace `pool.aquachain-foundation.org:8888` with the chosen pool and port configuration.

#### CPU

Pool mining with aquaminer command (now outdated, use for testing):

See https://github.com/aquanetwork/aquachain/issues/27

`./aquaminer.exe -r 1s -F http://pool.aquachain-foundation.org:8888/<address>/<worker>`

Here replace `<address>` with your wallet address and `<worker>` with any custom name for the CPU you are using with your address. Remember multiple workers can be used with a single wallet address and in this case all paid money will go to the same wallet.

`-t` flag for number of cpu (default all)

`./aquaminer.exe -r 1s -F http://pool.aquachain-foundation.org:8888/<address>/<worker>`

### Pools:

These are the currently known pools, add to this list if one is discovered!

  * http://pool.aquachain-foundation.org
  * https://aquacha.in
  * https://aqua.signal2noi.se/
  * https://aquapool.rplant.xyz


## Solo mining:

* **Don't solo mine unless you are running a pool, or have lots of rigs.**

* **Dont keep your keys on your RPC server**

* **Consider `-keystore dummykeys -aquabase 0x1234567abcdef12345600000` to mine to a specific address without having the key available.** 

Be sure to **wait and sync before mining**.  It doesn't take long.

To reduce orphan blocks, also be sure to **have peers** and check a block explorer to see the current block number and hash.

Do not key any keys inside the "nokeys" directory. You can safely delete `aquaminingdata` and `nokeys` (make sure you dont keep keys in there!)

**Run your RPC server like so: `aquachain -rpc -rpcaddr 192.168.1.5 -datadir aquaminingdata -keystore nokeys -aquabase 0x3317e8405e75551ec7eeeb3508650e7b349665ff`**

Later, to spend and use the AQUA console, just double click aquachain. This way, you keep your keys safe (in the default keystore dir) and don't mix `datadir`, this can prevent RPC attacks.

Please see the many cases where people have lost their ETH because leaving RPC open for even one minute.

### Solo farm

This assumes your AQUA node will be running from LAN 192.168.1.3, with other workers on the same lan.

WORKERS: `aquacppminer --solo -F http://192.168.1.3:8543/`

Also consider running a pool! [open-aquachain-pool](https://github.com/aquachain/open-aquachain-pool)

and see mining proxy: [aquachain-proxy](https://github.com/rplant8/aquachain-proxy)

**Coinbase**

use the `-aquabase` flag, or from console:

```
miner.setAquabase('0x3317e8405e75551ec7eeeb3508650e7b349665ff')
```

### CPU Benchmarks

See `#hashrate-reports` channel for many more: https://discord.gg/feEUajj
