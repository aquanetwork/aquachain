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
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

func TestNextHF(t *testing.T) {
	config := &ChainConfig{
		// simple hf map with forks at block 0, 21 and 30
		HF: ForkMap{
			0: big.NewInt(0),
			1: big.NewInt(21),
			2: big.NewInt(30),
		},
	}

	type test struct {
		input, expected *big.Int
	}

	tests := []test{
		{input: big.NewInt(10), expected: big.NewInt(21)},
		{input: big.NewInt(7), expected: big.NewInt(21)},
		{input: big.NewInt(11), expected: big.NewInt(21)},
		{input: big.NewInt(0), expected: big.NewInt(21)},
		{input: big.NewInt(21), expected: big.NewInt(30)},
		{input: big.NewInt(22), expected: big.NewInt(30)},
		{input: big.NewInt(23), expected: big.NewInt(30)},
		{input: big.NewInt(29), expected: big.NewInt(30)},
		{input: big.NewInt(30), expected: nil},
		{input: big.NewInt(35), expected: nil},
		{input: big.NewInt(350), expected: nil},
	}

	for i, test := range tests {
		output := config.NextHF(test.input)
		if test.expected == nil {
			if output == nil {
				continue
			} else {
				t.Errorf("Expected nil, got: %s", output)
				continue
			}
		}
		if output == nil && test.expected != nil {
			t.Errorf("Test %v failed.\nExpected: %s, Got nil", i, test.expected)
			continue
		}
		if output.Cmp(test.expected) != 0 {
			fmt.Printf("input %s errored\nGot: %s\nWanted:   %s\n", test.input, output, test.expected)
			t.Fail()
		}
	}
}
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
