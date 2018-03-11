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

package aquahash

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquanetwork/aquachain/common/math"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/params"
)

type diffTest struct {
	ParentTimestamp    uint64
	ParentDifficulty   *big.Int
	CurrentTimestamp   uint64
	CurrentBlocknumber *big.Int
	CurrentDifficulty  *big.Int
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	config := &params.ChainConfig{HomesteadBlock: big.NewInt(1)}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       new(big.Int).SetUint64(test.ParentTimestamp),
			Difficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
	}
}

func TestCalcDifficultyHF2(t *testing.T) {
	config := &params.ChainConfig{
		ChainId:        big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		EIP150Block:    big.NewInt(0),
		HF: map[int]*big.Int{
			0: big.NewInt(0),
			1: big.NewInt(1), // increase min difficulty to the next multiple of 2048
			2: big.NewInt(2), // increase min difficulty to the next multiple of 2048
		},
		Aquahash:    new(params.AquahashConfig),
		SupplyLimit: big.NewInt(42000000),
	}

	// 	ParentTimestamp    uint64
	// 	ParentDifficulty   *big.Int
	// 	CurrentTimestamp   uint64
	// 	CurrentBlocknumber *big.Int
	// 	CurrentDifficulty  *big.Int
	m := []diffTest{ // aiming for 240 second blocks
		//
		//
		//
		{000000001, big.NewInt(99999999), 80, big.NewInt(0), big.NewInt(99999999)},       // 1: frontier diff bug (not increasing)
		{000000001, big.NewInt(100001792), 80, big.NewInt(1), big.NewInt(100001792)},     // 2: hf1 diff bug (not increasing)
		{000000001, big.NewInt(100001792), 80, big.NewInt(2), big.NewInt(100050621)},     // 3: hf 2 (diff bug fixed)
		{000000001, big.NewInt(100050621), 100, big.NewInt(3), big.NewInt(100099473)},    // 4: 100 second block, should increase.
		{000000001, big.NewInt(100050621), 23, big.NewInt(4), big.NewInt(100099473)},     // 5: 23 second block, should increase
		{000000001, big.NewInt(100050621), 50, big.NewInt(5), big.NewInt(100099473)},     // 6: 50 second block, should increase
		{000000001, big.NewInt(100050621), 400, big.NewInt(6), big.NewInt(100001792)},    // 7: 400 second block, should decrease
		{000000001, big.NewInt(100050621), 200, big.NewInt(7), big.NewInt(100099473)},    // 8: 200 second block // should increase
		{000000001, big.NewInt(100050621), 235, big.NewInt(8), big.NewInt(100099473)},    // 9: 235 second block // should be increase
		{000000001, big.NewInt(100050621), 236, big.NewInt(8), big.NewInt(100099473)},    // 10: 236 second block // should be increase
		{000000001, big.NewInt(100050621), 237, big.NewInt(9), big.NewInt(100099473)},    // 237 second block // should be increase
		{000000001, big.NewInt(100050621), 238, big.NewInt(9), big.NewInt(100099473)},    // 238 second block // should be increase
		{000000001, big.NewInt(100050621), 245, big.NewInt(10), big.NewInt(100001792)},   // 245 second block// should decrease
		{000000001, big.NewInt(100050621), 255, big.NewInt(10), big.NewInt(100001792)},   // 255 second block// should decrease
		{000000001, big.NewInt(100050621), 320, big.NewInt(10), big.NewInt(100001792)},   // 255 second block//should decrease
		{000000001, big.NewInt(1000506210), 340, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 345, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 350, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 370, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 380, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 390, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 395, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 399, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
		{000000001, big.NewInt(1000506210), 500, big.NewInt(10), big.NewInt(1000017682)}, // 255 second block// should decrease
	}
	now := uint64(1520741293)
	for name, test := range m {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, now+test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       new(big.Int).SetUint64(now + test.ParentTimestamp),
			Difficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name+1, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
		//t.Logf("%v second block from %v = %v", test.CurrentTimestamp, test.CurrentDifficulty, diff)
		//fmt.Printf("%v %v %v %v\n", name, test.CurrentTimestamp, test.CurrentDifficulty, diff)
	}
}
