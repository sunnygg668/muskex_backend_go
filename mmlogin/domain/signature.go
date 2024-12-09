package domain

import (
	"github.com/ethereum/go-ethereum/common"
	"muskex/mmlogin/library/strutil"
)

type Signature [SignatureSize]byte

func newSignatureFromBytes(sigBytes []byte) Signature {
	sig := Signature{}
	copy(sig[:], sigBytes[:])

	if sig[SignatureSize-1] < SignatureRIRangeBase {
		sig[SignatureSize-1] += SignatureRIRangeBase
	}

	return sig
}

func NewSignatureFromHex(sigHex string) Signature {
	return newSignatureFromBytes(common.FromHex(sigHex))
}

func (sig Signature) Bytes() []byte {
	return sig[:]
}

func ValidateSignatureHex(sigHex string) error {
	if strutil.HasHexPrefix(sigHex) {
		sigHex = sigHex[2:]
	}

	if len(sigHex) != 2*SignatureSize {
		return ErrInvalidSignatureSize
	}
	if !strutil.IsHex(sigHex) {
		return ErrInvalidSignatureHex
	}
	return nil
}
