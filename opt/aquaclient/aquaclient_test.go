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

package aquaclient

import "gitlab.com/aquachain/aquachain"

// Verify that Client implements the aquachain interfaces.
var (
	_ = aquachain.ChainReader(&Client{})
	_ = aquachain.TransactionReader(&Client{})
	_ = aquachain.ChainStateReader(&Client{})
	_ = aquachain.ChainSyncReader(&Client{})
	_ = aquachain.ContractCaller(&Client{})
	_ = aquachain.GasEstimator(&Client{})
	_ = aquachain.GasPricer(&Client{})
	_ = aquachain.LogFilterer(&Client{})
	_ = aquachain.PendingStateReader(&Client{})
	// _ = aquachain.PendingStateEventer(&Client{})
	_ = aquachain.PendingContractCaller(&Client{})
)
