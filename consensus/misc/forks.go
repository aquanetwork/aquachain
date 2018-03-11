// Copyright 2017 The aquachain Authors
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
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/params"
)

// VerifyFork Doesn't do anything yet
func VerifyFork(config *params.ChainConfig, header *types.Header, uncle bool) error {
	// We don't care about uncles
	if uncle {
		return nil
	}
	// If the homestead reprice hash is set, validate it
	// for fork, number := range config.HF {
	// 	if number != nil && number.Cmp(header.Number) < 1 {
	// 		return nil
	// 	}
	// }
	// All ok, return
	return nil
}
