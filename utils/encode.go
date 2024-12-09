package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMd5FromBs(bs []byte) string {
	h := md5.New()
	h.Write(bs)
	return hex.EncodeToString(h.Sum(nil))
}
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
