// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !go1.10,amd64,!gccgo,!appengine

package argon2

func init() {
	useAVX2 = false
}

func processBlockAsm(out, in1, in2 *block, xor bool) {
	if useSSE4 {
		processBlockSSE4(out, in1, in2, xor)
	} else {
		processBlockSSE2(out, in1, in2, xor)
	}
}
