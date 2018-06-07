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

// Contains the metrics collected by the downloader.

package downloader

import (
	"gitlab.com/aquachain/aquachain/common/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("aqua/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("aqua/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("aqua/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("aqua/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("aqua/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("aqua/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("aqua/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("aqua/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("aqua/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("aqua/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("aqua/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("aqua/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("aqua/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("aqua/downloader/states/drop", nil)
)
