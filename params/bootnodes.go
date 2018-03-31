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
	// AquaChain Foundation Go Bootnodes
	"enode://7f636b8198a41abb10c1a571992335b8cb760d6ef973efc5f3ff613dda7acbe9e6d6b27254e076ef7b684ac7ea09a27bd05a37844cd8ad242199593bdd8cec21@107.161.24.142:21000", // aquachain-1
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21000", // aquachain-2
	"enode://7f636b8198a41abb10c1a571992335b8cb760d6ef973efc5f3ff613dda7acbe9e6d6b27254e076ef7b684ac7ea09a27bd05a37844cd8ad242199593bdd8cec21@107.161.24.142:21001", // aquachain-1 new protocol
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21001", // aquachain-2 new protocol
	"enode://1a6b78cf626540d1eecfeba1f364e72bf92847561b9344403ac7010b2be184cfc760b5bcd21402b19713deebef256dcdfc5af67487554650bf07807737a36203@23.94.123.137:21303",  // aerthnode
	"enode://a341920437d7141e4355ed1f298fd2415cbec781c8b4fedd943eac37fd0c835375718085b1a65208c0a06af10c388d452a4148b8430da93bd0b75100b2315f3c@107.161.24.142:21303", // aerthpool
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// test network.
var TestnetBootnodes = []string{
	"enode://6227ff2948ff51ee4f09e5f1df2c1270c47b753718d406605787326341de6ff8e7cb6a5f01a4deed5437dcdd7b9fb8e656f0ad6a08c1f677c2ca5d0e666a92fc@168.235.78.103:21002", // aquachain-2 testnet new protocol
}

// RinkebyBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Rinkeby test network.
var RinkebyBootnodes = []string{}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{}
