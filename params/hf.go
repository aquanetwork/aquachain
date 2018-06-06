// Copyright 2016 The aquachain Authors
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
	for i := 0; i < len(f); i++ {
		s = fmt.Sprintf("%s %v:%v", s, i, f[i].Int64())
	}
	return strings.TrimSpace(s)
}

type HeaderVersion byte

func (c ChainConfig) GetBlockVersion(height *big.Int) HeaderVersion {
	if height == nil {
		return 2
	}
	if height.Uint64() != 0 && c.IsHF(5, height) {
		return 2
	}
	return 1
}
