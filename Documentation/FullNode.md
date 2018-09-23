# how to run an archive node

1. If synchronized, use `aquachain export mainnet.dat` and then `aquachain removedb`
2. Run `aquachain -gcmode archive import mainnet.dat`
3. Now, run aquachain with `-gcmode archive` flag.