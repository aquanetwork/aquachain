// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// +build !gccgo,!cgo

package tracers

import (
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/core/vm"
)

// Tracer provides an implementation of Tracer that evaluates a Javascript
// function for each VM execution step.
type Tracer struct{}

var ErrNoCgo = errors.New("This version of aquachain has been compiled with pure go, and has no tracers.")

// New instantiates a new tracer instance. code specifies a Javascript snippet,
// which must evaluate to an expression returning an object with 'step', 'fault'
// and 'result' functions.
func New(code string) (*Tracer, error) {
	return nil, ErrNoCgo
}

// Stop terminates execution of the tracer at the first opportune moment.
func (jst *Tracer) Stop(error) {}

func (jst *Tracer) CaptureStart(from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) error {
	return ErrNoCgo
}

// CaptureState implements the Tracer interface to trace a single step of VM execution.
func (jst *Tracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas,
	cost uint64, memory *vm.Memory, stack *vm.Stack,
	contract *vm.Contract, depth int, err error) error {
	return ErrNoCgo
}

// CaptureFault implements the Tracer interface to trace an execution fault
// while running an opcode.
func (jst *Tracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, contract *vm.Contract, depth int, err error) error {
	return ErrNoCgo
}

// CaptureEnd is called after the call finishes to finalize the tracing.
func (jst *Tracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	return ErrNoCgo
}

// GetResult calls the Javascript 'result' function and returns its value, or any accumulated error
func (jst *Tracer) GetResult() (json.RawMessage, error) {
	// Transform the context into a JavaScript object and inject into the state
	return json.RawMessage{}, ErrNoCgo
}
