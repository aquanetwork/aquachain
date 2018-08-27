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

package params

import (
	"math/big"
	"reflect"
	"testing"
)

func TestCheckCompatible(t *testing.T) {
	type test struct {
		stored, new *ChainConfig
		head        uint64
		wantErr     *ConfigCompatError
	}
	tests := []test{
		{stored: AllAquahashProtocolChanges, new: AllAquahashProtocolChanges, head: 0, wantErr: nil},
		{stored: AllAquahashProtocolChanges, new: AllAquahashProtocolChanges, head: 100, wantErr: nil},
		{
			stored:  &ChainConfig{EIP150Block: big.NewInt(10)},
			new:     &ChainConfig{EIP150Block: big.NewInt(20)},
			head:    9,
			wantErr: nil,
		},
		{
			stored: AllAquahashProtocolChanges,
			new:    &ChainConfig{HomesteadBlock: nil},
			head:   3,
			wantErr: &ConfigCompatError{
				What:         "Homestead fork block",
				StoredConfig: big.NewInt(0),
				NewConfig:    nil,
				RewindTo:     0,
			},
		},
		{
			stored: AllAquahashProtocolChanges,
			new:    &ChainConfig{HomesteadBlock: big.NewInt(1)},
			head:   3,
			wantErr: &ConfigCompatError{
				What:         "Homestead fork block",
				StoredConfig: big.NewInt(0),
				NewConfig:    big.NewInt(1),
				RewindTo:     0,
			},
		},
		{
			stored: &ChainConfig{HomesteadBlock: big.NewInt(30), EIP150Block: big.NewInt(10)},
			new:    &ChainConfig{HomesteadBlock: big.NewInt(25), EIP150Block: big.NewInt(20)},
			head:   25,
			wantErr: &ConfigCompatError{
				What:         "EIP150 fork block",
				StoredConfig: big.NewInt(10),
				NewConfig:    big.NewInt(20),
				RewindTo:     9,
			},
		},
	}

	for _, test := range tests {
		err := test.stored.CheckCompatible(test.new, test.head)
		if !reflect.DeepEqual(err, test.wantErr) {
			t.Errorf("error mismatch:\nstored: %v\nnew: %v\nhead: %v\nerr: %v\nwant: %v", test.stored, test.new, test.head, err, test.wantErr)
		}
	}
}

func TestNextHF(t *testing.T) {
	type tc struct {
		number *big.Int
		hf     int
		gold   *big.Int
	}
	for i, config := range []*ChainConfig{MainnetChainConfig} {
		for _, v := range []tc{
			{big.NewInt(0), 0, config.HF[0]},
			{big.NewInt(1), 0, config.HF[0]},
			{big.NewInt(3001), 1, config.HF[1]},
			{big.NewInt(3600), 2, config.HF[2]},
			{big.NewInt(64000), 9, config.HF[9]},
			{big.NewInt(84000000), 0, nil},
		} {
			gothf, got := config.NextHF(v.number)
			if got == nil {
				if v.gold == nil {
					continue
				}
				t.Logf("test %v: wanted %s, got nil", i, v.gold)
				t.Fail()
				continue
			}
			if gothf != v.hf {
				t.Logf("test %v: got %v, expected %v", i, gothf, v.hf)
				t.Fail()
			}

			if v.gold == nil || got.Cmp(v.gold) != 0 {
				t.Logf("test %v: got %s, expected %s", i, got, v.gold)
				t.Fail()
			}
		}
	}
}
