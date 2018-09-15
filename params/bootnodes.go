// Copyright 2015 The aquachain Authors
// This file is part of the aquachain library.
//
// The aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aquachain library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main AquaChain network.
var MainnetBootnodes = []string{
	// AquaChain Foundation Go Bootnodes (port 21000)
	"enode://7f636b8198a41abb10c1a571992335b8cb760d6ef973efc5f3ff613dda7acbe9e6d6b27254e076ef7b684ac7ea09a27bd05a37844cd8ad242199593bdd8cec21@107.161.24.142:21000", // aquachain-1
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21000", // aquachain-2
	"enode://1a6b78cf626540d1eecfeba1f364e72bf92847561b9344403ac7010b2be184cfc760b5bcd21402b19713deebef256dcdfc5af67487554650bf07807737a36203@23.94.123.137:21303",  // aerthnode
	"enode://a341920437d7141e4355ed1f298fd2415cbec781c8b4fedd943eac37fd0c835375718085b1a65208c0a06af10c388d452a4148b8430da93bd0b75100b2315f3c@107.161.24.142:21303", // aerthpool
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network. (port 21001)
var DiscoveryV5Bootnodes = []string{
	"enode://7f636b8198a41abb10c1a571992335b8cb760d6ef973efc5f3ff613dda7acbe9e6d6b27254e076ef7b684ac7ea09a27bd05a37844cd8ad242199593bdd8cec21@107.161.24.142:21001", // aquachain-1 new protocol
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21001", // aquachain-2 new protocol
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// test network. (port 21002)
var TestnetBootnodes = []string{
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21002", // aquachain-2 testnet new protocol
}

// Testnet2Bootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Testnet2 test network.
var Testnet2Bootnodes = []string{}

var EthnetBootnodes = []string{
	// Ethereum Foundation Go Bootnodes
	"enode://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303", // IE
	"enode://3f1d12044546b76342d59d4a05532c14b85aa669704bfe1f864fe079415aa2c02d743e03218e57a33fb94523adb54032871a6c51b2cc5514cb7c7e35b3ed0a99@13.93.211.84:30303",  // US-WEST
	"enode://78de8a0916848093c73790ead81d1928bec737d565119932b98c6b100d944b7a95e94f847f689fc723399d2e31129d182f7ef3863f2b4c820abbf3ab2722344d@191.235.84.50:30303", // BR
	"enode://158f8aab45f6d19c6cbf4a089c2670541a8da11978a2f90dbf6a502a4a3bab80d288afdbeb7ec0ef6d92de563767f3b1ea9e8e334ca711e9f8e2df5a0385e8e6@13.75.154.138:30303", // AU
	"enode://1118980bf48b0a3640bdba04e0fe78b1add18e1cd99bf22d53daac1fd9972ad650df52176e7c7d89d1114cfef2bc23a2959aa54998a46afcf7d91809f0855082@52.74.57.123:30303",  // SG

	// Ethereum Foundation C++ Bootnodes
	"enode://979b7fa28feeb35a4741660a16076f1943202cb72b6af70d327f053e248bab9ba81760f39d0701ef1d8f89cc1fbd2cacba0710a12cd5314d5e0c9021aa3637f9@5.1.83.226:30303", // DE
}
