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
	for i := 0; i < 10; i++ {
		if f[i] == nil {
			continue
		}
		s = fmt.Sprintf("%s %v:%v", s, i, f[i].Int64())
	}
	return strings.TrimSpace(s)
}

type HeaderVersion byte

func (c ChainConfig) GetBlockVersion(height *big.Int) HeaderVersion {
	if height == nil {
		panic("chainconfig: nil height, no block version")
	}

	var (
		h = height.Uint64()
	)

	if h != 0 && c.IsHF(9, height) && h%2 == 0 {
		return 4 // argon2id, 1, 512, 1
	}

	if h != 0 && c.IsHF(9, height) {
		return 3 // argon2id, 1, 256, 1
	}

	if h != 0 && c.IsHF(8, height) {
		return 3 // argon2id, 1, 256, 1
	}

	if h != 0 && c.IsHF(5, height) { // argon2id, 1, 1, 1
		return 2
	}

	return 1 // ethash
}
