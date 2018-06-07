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

package misc

import (
	"math/big"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/core/state"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/params"
)

func VerifyHFHeaderExtraData(config *params.ChainConfig, header *types.Header) error {
	return nil
}

// ApplyHardFork modifies the state database according to the specific hf
func ApplyHardFork(statedb *state.StateDB) {
	// do nothing
}

func ApplyHardFork4(statedb *state.StateDB) {
	big0 := new(big.Int)
	for _, hexadd := range DeallocListHF4 {
		address := common.HexToAddress(hexadd)
		if statedb.Exist(address) {
			statedb.SetBalance(address, big0)
		}
	}
}

func ApplyHardFork5(statedb *state.StateDB) {
	for _, hexadd := range DeallocListHF4 {
		address := common.HexToAddress(hexadd)
		if statedb.Exist(address) {
			statedb.Empty(address)
		}
	}
}
