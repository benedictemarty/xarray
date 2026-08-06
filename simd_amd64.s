//go:build amd64 && !noasm

#include "textflag.h"

// func addFloat64AVX(dst, x, y *float64, n int)
// Additionne n float64 (n multiple de 4) : dst[i] = x[i] + y[i].
// Utilise VADDPD sur registres YMM (4 float64 par itération).
TEXT ·addFloat64AVX(SB), NOSPLIT, $0-32
	MOVQ dst+0(FP), DI
	MOVQ x+8(FP), SI
	MOVQ y+16(FP), DX
	MOVQ n+24(FP), CX
	XORQ AX, AX

	// Boucle déroulée ×4 : 16 float64 par itération, 4 registres YMM
	// indépendants (meilleur parallélisme d'instructions).
loop16:
	MOVQ CX, BX
	SUBQ AX, BX
	CMPQ BX, $16
	JL   loop4
	VMOVUPD (SI)(AX*8), Y0
	VADDPD  (DX)(AX*8), Y0, Y0
	VMOVUPD Y0, (DI)(AX*8)
	VMOVUPD 32(SI)(AX*8), Y1
	VADDPD  32(DX)(AX*8), Y1, Y1
	VMOVUPD Y1, 32(DI)(AX*8)
	VMOVUPD 64(SI)(AX*8), Y2
	VADDPD  64(DX)(AX*8), Y2, Y2
	VMOVUPD Y2, 64(DI)(AX*8)
	VMOVUPD 96(SI)(AX*8), Y3
	VADDPD  96(DX)(AX*8), Y3, Y3
	VMOVUPD Y3, 96(DI)(AX*8)
	ADDQ    $16, AX
	JMP     loop16

	// Reste : 4 float64 par itération.
loop4:
	MOVQ CX, BX
	SUBQ AX, BX
	CMPQ BX, $4
	JL   done
	VMOVUPD (SI)(AX*8), Y0
	VADDPD  (DX)(AX*8), Y0, Y0
	VMOVUPD Y0, (DI)(AX*8)
	ADDQ    $4, AX
	JMP     loop4

done:
	VZEROUPPER
	RET

// func cpuHasAVX() bool
// Vrai si le processeur ET le système d'exploitation supportent AVX :
//   - CPUID(1) : ECX bit 27 (OSXSAVE) et bit 28 (AVX) ;
//   - XGETBV(0) : XCR0 bits 1 et 2 (état SSE + YMM sauvegardé par l'OS).
TEXT ·cpuHasAVX(SB), NOSPLIT, $0-1
	MOVQ $1, AX
	CPUID
	MOVL CX, DX
	ANDL $0x18000000, DX      // bits 27 (OSXSAVE) + 28 (AVX)
	CMPL DX, $0x18000000
	JNE  no

	MOVL   $0, CX
	XGETBV                    // -> EDX:EAX = XCR0
	ANDL   $6, AX             // bits 1 (SSE) + 2 (YMM)
	CMPL   AX, $6
	JNE    no

	MOVB $1, ret+0(FP)
	RET

no:
	MOVB $0, ret+0(FP)
	RET
