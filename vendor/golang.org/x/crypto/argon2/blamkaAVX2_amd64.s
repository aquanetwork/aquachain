// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build go1.10,amd64,!gccgo,!appengine

#include "textflag.h"

DATA ·AVX2_c40<>+0x00(SB)/8, $0x0201000706050403
DATA ·AVX2_c40<>+0x08(SB)/8, $0x0a09080f0e0d0c0b
DATA ·AVX2_c40<>+0x10(SB)/8, $0x0201000706050403
DATA ·AVX2_c40<>+0x18(SB)/8, $0x0a09080f0e0d0c0b
GLOBL ·AVX2_c40<>(SB), (NOPTR+RODATA), $32

DATA ·AVX2_c48<>+0x00(SB)/8, $0x0100070605040302
DATA ·AVX2_c48<>+0x08(SB)/8, $0x09080f0e0d0c0b0a
DATA ·AVX2_c48<>+0x10(SB)/8, $0x0100070605040302
DATA ·AVX2_c48<>+0x18(SB)/8, $0x09080f0e0d0c0b0a
GLOBL ·AVX2_c48<>(SB), (NOPTR+RODATA), $32

#define SHUFFLE(v1, v2, v3) \
	VPERMQ $0x39, v1, v1; \
	VPERMQ $0x4E, v2, v2; \
	VPERMQ $-109, v3, v3

#define HALF_ROUND(v0, v1, v2, v3, t0, t1, c40, c48) \
	VPMULUDQ v0, v1, t0;    \
	VPADDQ   v0, t0, t1;    \
	VPADDQ   v1, t0, t0;    \
	VPADDQ   t1, t0, v0;    \
	VPXOR    v0, v3, v3;    \
	VPSHUFD  $0xB1, v3, v3; \
	VPMULUDQ v2, v3, t0;    \
	VPADDQ   v2, t0, t1;    \
	VPADDQ   v3, t0, t0;    \
	VPADDQ   t1, t0, v2;    \
	VPXOR    v2, v1, v1;    \
	VPSHUFB  c40, v1, v1;   \
	VPMULUDQ v0, v1, t0;    \
	VPADDQ   v0, t0, t1;    \
	VPADDQ   v1, t0, t0;    \
	VPADDQ   t1, t0, v0;    \
	VPXOR    v0, v3, v3;    \
	VPSHUFB  c48, v3, v3;   \
	VPMULUDQ v2, v3, t0;    \
	VPADDQ   v2, t0, t1;    \
	VPADDQ   v3, t0, t0;    \
	VPADDQ   t1, t0, v2;    \
	VPXOR    v2, v1, v1;    \
	VPSLLQ   $1, v1, t0;    \
	VPSRLQ   $63, v1, v1;   \
	VPXOR    t0, v1, v1

#define LOAD_MSG_0(block, off) \
	VMOVDQU 8*(off+0)(block), Y0; \
	VMOVDQU 8*(off+4)(block), Y1; \
	VMOVDQU 8*(off+8)(block), Y2; \
	VMOVDQU 8*(off+12)(block), Y3

#define STORE_MSG_0(block, off) \
	VMOVDQU Y0, 8*(off+0)(block); \
	VMOVDQU Y1, 8*(off+4)(block); \
	VMOVDQU Y2, 8*(off+8)(block); \
	VMOVDQU Y3, 8*(off+12)(block)

#define PREFETCH_MSG_1(block, off, t0, t1, t2, t3, t4, t5, t6, t7) \
	VMOVDQU 8*off+0*8(block), t0;  \
	VMOVDQU 8*off+16*8(block), t1; \
	VMOVDQU 8*off+32*8(block), t2; \
	VMOVDQU 8*off+48*8(block), t3; \
	VMOVDQU 8*off+64*8(block), t4; \
	VMOVDQU 8*off+80*8(block), t5; \
	VMOVDQU 8*off+96*8(block), t6; \
	VMOVDQU 8*off+112*8(block), t7

#define LOAD_MSG_1_F(t0, t1, t2, t3, t4, t5, t6, t7) \
	VPERM2I128 $0x20, t1, t0, Y0; \
	VPERM2I128 $0x20, t3, t2, Y1; \
	VPERM2I128 $0x20, t5, t4, Y2; \
	VPERM2I128 $0x20, t7, t6, Y3

#define LOAD_MSG_1_S(t0, t1, t2, t3, t4, t5, t6, t7) \
	VPERM2I128 $0x31, t1, t0, Y0; \
	VPERM2I128 $0x31, t3, t2, Y1; \
	VPERM2I128 $0x31, t5, t4, Y2; \
	VPERM2I128 $0x31, t7, t6, Y3

#define STORE_MSG_1(block, off) \
	VEXTRACTI128 $0, Y0, 8*off+0*8(block);  \
	VEXTRACTI128 $1, Y0, 8*off+16*8(block); \
	VEXTRACTI128 $0, Y1, 8*off+32*8(block); \
	VEXTRACTI128 $1, Y1, 8*off+48*8(block); \
	VEXTRACTI128 $0, Y2, 8*off+64*8(block); \
	VEXTRACTI128 $1, Y2, 8*off+80*8(block); \
	VEXTRACTI128 $0, Y3, 8*off+96*8(block); \
	VEXTRACTI128 $1, Y3, 8*off+112*8(block)

#define BLAMKA_ROUND_0(block, off, v0, v1, v2, v3, t0, t1, c40, c48) \
	LOAD_MSG_0(block, off);                       \
	HALF_ROUND(v0, v1, v2, v3, t0, t1, c40, c48); \
	SHUFFLE(Y1, Y2, Y3);                          \
	HALF_ROUND(v0, v1, v2, v3, t0, t1, c40, c48); \
	SHUFFLE(Y3, Y2, Y1);                          \
	STORE_MSG_0(block, off)

#define BLAMKA_ROUND_1(block, off, v0, v1, v2, v3, t0, t1, c40, c48, LOAD_MSG) \
	LOAD_MSG(Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15); \
	HALF_ROUND(v0, v1, v2, v3, t0, t1,  c40, c48);  \
	SHUFFLE(Y1, Y2, Y3);                            \
	HALF_ROUND(v0, v1, v2, v3, t0, t1, c40, c48);   \
	SHUFFLE(Y3, Y2, Y1);                            \
	STORE_MSG_1(block, off)

// func blamkaAVX2(b *block)
TEXT ·blamkaAVX2(SB), 4, $0-8
	MOVQ b+0(FP), AX

	VMOVDQU ·AVX2_c40<>(SB), Y6
	VMOVDQU ·AVX2_c48<>(SB), Y7

	BLAMKA_ROUND_0(AX, 0, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 16, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 32, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 48, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 64, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 80, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 96, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)
	BLAMKA_ROUND_0(AX, 112, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7)

	PREFETCH_MSG_1(AX, 0, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15)
	BLAMKA_ROUND_1(AX, 0, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_F)
	BLAMKA_ROUND_1(AX, 2, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_S)
	PREFETCH_MSG_1(AX, 4, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15)
	BLAMKA_ROUND_1(AX, 4, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_F)
	BLAMKA_ROUND_1(AX, 6, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_S)
	PREFETCH_MSG_1(AX, 8, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15)
	BLAMKA_ROUND_1(AX, 8, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_F)
	BLAMKA_ROUND_1(AX, 10, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_S)
	PREFETCH_MSG_1(AX, 12, Y8, Y9, Y10, Y11, Y12, Y13, Y14, Y15)
	BLAMKA_ROUND_1(AX, 12, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_F)
	BLAMKA_ROUND_1(AX, 14, Y0, Y1, Y2, Y3, Y4, Y5, Y6, Y7, LOAD_MSG_1_S)

	VZEROUPPER
	RET

#define MIX(dst, off, a, b, c) \
	VMOVDQU 32*off(a), Y0;  \
	VMOVDQU 32*off(b), Y1;  \
	VMOVDQU 32*off(c), Y2;  \
	VPXOR   Y0, Y1, Y3;     \
	VPXOR   Y2, Y3, Y0;     \
	VMOVDQU Y0, 32*off(dst)

// func mixBlocksAVX2(out, a, b, c *block)
TEXT ·mixBlocksAVX2(SB), 4, $0-32
	MOVQ out+0(FP), DI
	MOVQ a+8(FP), AX
	MOVQ b+16(FP), BX
	MOVQ a+24(FP), CX

	MIX(DI, 0, AX, BX, CX)
	MIX(DI, 1, AX, BX, CX)
	MIX(DI, 2, AX, BX, CX)
	MIX(DI, 3, AX, BX, CX)
	MIX(DI, 4, AX, BX, CX)
	MIX(DI, 5, AX, BX, CX)
	MIX(DI, 6, AX, BX, CX)
	MIX(DI, 7, AX, BX, CX)

	MIX(DI, 8, AX, BX, CX)
	MIX(DI, 9, AX, BX, CX)
	MIX(DI, 10, AX, BX, CX)
	MIX(DI, 11, AX, BX, CX)
	MIX(DI, 12, AX, BX, CX)
	MIX(DI, 13, AX, BX, CX)
	MIX(DI, 14, AX, BX, CX)
	MIX(DI, 15, AX, BX, CX)

	MIX(DI, 16, AX, BX, CX)
	MIX(DI, 17, AX, BX, CX)
	MIX(DI, 18, AX, BX, CX)
	MIX(DI, 19, AX, BX, CX)
	MIX(DI, 20, AX, BX, CX)
	MIX(DI, 21, AX, BX, CX)
	MIX(DI, 22, AX, BX, CX)
	MIX(DI, 23, AX, BX, CX)

	MIX(DI, 24, AX, BX, CX)
	MIX(DI, 25, AX, BX, CX)
	MIX(DI, 26, AX, BX, CX)
	MIX(DI, 27, AX, BX, CX)
	MIX(DI, 28, AX, BX, CX)
	MIX(DI, 29, AX, BX, CX)
	MIX(DI, 30, AX, BX, CX)
	MIX(DI, 31, AX, BX, CX)

	VZEROUPPER
	RET

#define XOR(dst, off, a, b, c) \
	VMOVDQU 32*off(a), Y0;   \
	VMOVDQU 32*off(b), Y1;   \
	VMOVDQU 32*off(c), Y2;   \
	VMOVDQU 32*off(dst), Y3; \
	VPXOR   Y0, Y1, Y4;      \
	VPXOR   Y2, Y3, Y5;      \
	VPXOR   Y4, Y5, Y6;      \
	VMOVDQU Y6, 32*off(DI)

// func xorBlocksAVX2(out, a, b, c *block)
TEXT ·xorBlocksAVX2(SB), 4, $0-32
	MOVQ out+0(FP), DI
	MOVQ a+8(FP), AX
	MOVQ b+16(FP), BX
	MOVQ a+24(FP), CX

	XOR(DI, 0, AX, BX, CX)
	XOR(DI, 1, AX, BX, CX)
	XOR(DI, 2, AX, BX, CX)
	XOR(DI, 3, AX, BX, CX)
	XOR(DI, 4, AX, BX, CX)
	XOR(DI, 5, AX, BX, CX)
	XOR(DI, 6, AX, BX, CX)
	XOR(DI, 7, AX, BX, CX)

	XOR(DI, 8, AX, BX, CX)
	XOR(DI, 9, AX, BX, CX)
	XOR(DI, 10, AX, BX, CX)
	XOR(DI, 11, AX, BX, CX)
	XOR(DI, 12, AX, BX, CX)
	XOR(DI, 13, AX, BX, CX)
	XOR(DI, 14, AX, BX, CX)
	XOR(DI, 15, AX, BX, CX)

	XOR(DI, 16, AX, BX, CX)
	XOR(DI, 17, AX, BX, CX)
	XOR(DI, 18, AX, BX, CX)
	XOR(DI, 19, AX, BX, CX)
	XOR(DI, 20, AX, BX, CX)
	XOR(DI, 21, AX, BX, CX)
	XOR(DI, 22, AX, BX, CX)
	XOR(DI, 23, AX, BX, CX)

	XOR(DI, 24, AX, BX, CX)
	XOR(DI, 25, AX, BX, CX)
	XOR(DI, 26, AX, BX, CX)
	XOR(DI, 27, AX, BX, CX)
	XOR(DI, 28, AX, BX, CX)
	XOR(DI, 29, AX, BX, CX)
	XOR(DI, 30, AX, BX, CX)
	XOR(DI, 31, AX, BX, CX)

	VZEROUPPER
	RET
