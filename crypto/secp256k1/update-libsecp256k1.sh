#!/bin/bash
set -e
test -d libsecp256k1 || (echo "please run this script inside the /crypto/secp256k1 directory." && exit 111)
set -x
rm -rv libsecp256k1
git clone https://github.com/bitcoin-core/secp256k1 libsecp256k1
rm -rfv libsecp256k1/.git
set +x
echo libsecp256k1 upgraded
