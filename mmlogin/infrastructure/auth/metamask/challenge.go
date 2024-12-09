package metamask

import (
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	challengeType = "string"
	challengeName = "challenge"
)

type challenge string

func (chal challenge) signatureHashBytes() []byte {
	data := append([]byte("\u0019TRON Signed Message:\n32"), []byte(chal)...)
	sighash := crypto.Keccak256(data)
	return sighash
}
