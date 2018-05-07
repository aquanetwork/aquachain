// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.10,amd64,!gccgo,!appengine

package argon2

import "golang.org/x/sys/cpu"

func init() {
	useAVX2 = cpu.X86.HasAVX2
}

// This function is implemented in blamkaAVX2_amd64.s
//go:noescape
func mixBlocksAVX2(out, a, b, c *block)

// This function is implemented in blamkaAVX2_amd64.s
//go:noescape
func xorBlocksAVX2(out, a, b, c *block)

// This function is implemented in blamkaAVX2_amd64.s
//go:noescape
func blamkaAVX2(b *block)

func processBlockAsm(out, in1, in2 *block, xor bool) {
	switch {
	case useAVX2:
		processBlockAVX2(out, in1, in2, xor)
	case useSSE4:
		processBlockSSE4(out, in1, in2, xor)
	default:
		processBlockSSE2(out, in1, in2, xor)
	}
}

func processBlockAVX2(out, in1, in2 *block, xor bool) {
	var t block
	mixBlocksAVX2(&t, in1, in2, &t)
	blamkaAVX2(&t)
	if xor {
		xorBlocksAVX2(out, in1, in2, &t)
	} else {
		mixBlocksAVX2(out, in1, in2, &t)
	}
}
