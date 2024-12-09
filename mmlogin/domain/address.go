package domain

import (
	"errors"
	tronAddress "github.com/fbsobreira/gotron-sdk/pkg/address"
)

type Address tronAddress.Address

func NewAddressFromHex(addressHex string) Address {
	address, _ := tronAddress.Base58ToAddress(addressHex)
	return Address(address)
}

func (address Address) Hex() string {
	return tronAddress.Address(address).Hex()
}

func ValidateAddressHex(addressHex string) error {
	bs, err := tronAddress.Base58ToAddress(addressHex)
	if err != nil {
		return errors.New("err base58 format")
	}
	if len(bs) != 21 {
		return errors.New("err addres len")
	}
	if bs[0] != 0x41 {
		return errors.New("err format")
	}
	return nil
}
