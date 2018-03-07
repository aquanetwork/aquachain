# Aquachain

## General Purpose Smart Contract Distributed Super Computer

Only one AQUA reward every block. Avg block time is 197 seconds as of right now*. At block 42mil block rewards become zero (gas only)

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

Example command line for mining and using a LOCALLY stored copy of MEW AND the block explorer(found at http://explorer.aquanetwork.co):

```
aquachain --mine --rpc --rpccorsdomain='*' --rpcvhosts='*' --rpcapi 'eth,aqua,web3' console
```

This will listen on port 8543, so make sure to use 'http://localhost' and '8543' when setting up MEW custom node.

## Resources

Explorer - http://explorer.aquanetwork.co/
Downloads - http://explorer.aquanetwork.co/dl/
Github - http://github.com/aquachain
Github - http://github.com/aquanetwork/aquachain
News, Chat: https://t.me/AquaCrypto


Check your downloads, latest release will be edited from here.

```
MD5
d7e943fd613554a7be52614ac6c2eccf  ./aquachain-linux64
a7d1c9fd7f42e9e88a11dc78056b2994  ./aquachain-1.0.exe
6f4c576d268169bc3e17277e71eec434  ./aquachain-linux32
e16fe29cc16824fe2183b73270ccb697  ./aquachain-mac-osx
bfeec6e0ca7c221df6b4d008a4bccc88  ./all/aquad-darwin-10.6-amd64.zip
bd8f0320fe7f0dd9c1cedf1fbe1c1b28  ./all/aquad-windows-4.0-amd64.exe.zip
34d9e48fa878242660a94af4d758a290  ./all/sha256sum.txt
949ce56b83b6a9b92b6a0831d7a783d2  ./all/aquad-linux-amd64.zip
300e5b306be118fd17fb27e5fe139edc  ./all/aqua-explorer.zip
5e6de1649b9bc89036e86ec53c871143  ./all/bootstrap-small.dat.zip
8e792d6fcee41203f87d73a11602517a  ./all/aquawallet-v3.20.01.zip
c7b5220f7c94eca6484b73d076f9d8b6  ./all/bootstrap.dat
554202723277cebe3795a7daa8c7e21c  ./all/aquad-linux-386.zip
a6d878a781693be56f89d72e2ae5bd07  ./all/bootstrap.dat.zip
c7e2e0dc11f7f19c3782905ef7f717aa  ./all/aquawallet-v3.20.01.tar.gz
6b85f0bf06143571adfdb0a0531651bf  ./all/aquachain-1.0.0.tar.gz
f63932cad00a43286753d58d810a2965  ./all/bootstrap-small.dat

SHA256
30f8923879478fe159fc8d9366f2d027819c9f04110a6e3382dae37ab6a804b9  ./aquachain-linux64
5c775b6394566bdb7b2a1758adb515a946cfa541dfc423b9215677b41d1eef17  ./aquachain-1.0.exe
a48bc273b3629962cf523ca09c18f2e99fde8237e43127aae523f9d7969c2b68  ./aquachain-linux32
a3e08de687caf54fc659acd55d1417dff686064e16ccd3a43fd0d14ad919dd00  ./aquachain-mac-osx
e35a54eeba4e8bbc2ff6f97e383b690fca4406a1f388a9fba698745e9b9e6bbe  ./all/aquad-darwin-10.6-amd64.zip
7818879654aacedc667662d466dc463c1bc257982c715fe9f03b3e934b649aa3  ./all/aquad-windows-4.0-amd64.exe.zip
8d7662047f63c79cfd7c481dc98b2933d16bcfdd0072254459e97da4e680138b  ./all/sha256sum.txt
6683e2d44bca62d798324d54a86520b5cc6e59475063f3fc49ce5427379a5f4b  ./all/aquad-linux-amd64.zip
c78193ba66af30ac8ab3ff5592473fad14f98cb2eb63b9c803e0147b8ff7052a  ./all/aqua-explorer.zip
47e43d641ecfd4fff649ead4a22ae967127ba956d1e2d9dfa1b124cb67421761  ./all/bootstrap-small.dat.zip
2f199099d76da2bbc838bbe12ac908e7c576ee9839668b79d65def475690a2db  ./all/aquawallet-v3.20.01.zip
224300de0cb3f34395750c538de8ba50d90c9e646a398f0df369afd89b7fce21  ./all/bootstrap.dat
6512498eefc074831d9595e245f183164ea719036bef3e2be029c66cc2675378  ./all/aquad-linux-386.zip
0ad9694e97b9ff87db3722ee67c08ca0196b325ac029c741d1d243589db18b57  ./all/bootstrap.dat.zip
8f75ec7e66b231a50b54810303ed127abe38c7afe9feb871a8a0813d298d33ec  ./all/aquawallet-v3.20.01.tar.gz
5d11f1620fbeebf369aa743570bb14387b26d3247693ee2e632aa1e80aab0c45  ./all/aquachain-1.0.0.tar.gz
2574d824703ee72986305a63953a426033e1ef6ff319dc838064d7f25fd65f16  ./all/bootstrap-small.dat

```

Contributions welcome. Check out @AquaCrypto on telegram for ways to help.

