// Copyright 2018 The aquachain Authors
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

import (
	"fmt"
	"math/big"
	"strings"
)

type ForkMap map[int]*big.Int

func (f ForkMap) String() (s string) {
	for i := 0; i <= KnownHF; i++ {
		if f[i] == nil {
			continue
		}
		s = fmt.Sprintf("%s %v:%v", s, i, f[i].Int64())
	}
	return strings.TrimSpace(s)
}

// HeaderVersion is not stored in db, or rlp encoded, or sent over the network.
type HeaderVersion byte

func (c ChainConfig) GetBlockVersion(height *big.Int) HeaderVersion {
	if height == nil {
		panic("GetBlockVersion: got nil height")
	}
	if c.IsHF(8, height) {
		return 3 // argon2id (1,256mb,1)
	}
	if c.IsHF(5, height) {
		return 2 // argon2id (1,1kb,1)
	}
	return 1 // ethash
}

// IsHF returns whether num is either equal to the hf block or greater.
func (c *ChainConfig) IsHF(hf int, num *big.Int) bool {
	if c.HF[hf] == nil {
		return false
	}
	return isForked(c.HF[hf], num)
}

// GetHF returns the height of input hf, can be nil.
func (c *ChainConfig) GetHF(hf int) *big.Int {
	if c.HF[hf] == nil {
		return nil
	}
	return new(big.Int).Set(c.HF[hf])
}

// NextHF returns the next scheduled hard fork block number
func (c *ChainConfig) NextHF(cur *big.Int) *big.Int {
	for i := KnownHF; i > 0; i-- {
		if c.HF[i] == nil {
			continue
		}
		return new(big.Int).Set(c.HF[i])
	}
	return nil

}
