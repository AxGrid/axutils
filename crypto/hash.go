package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GetMD5HashBytes(text string) []byte {
	hash := md5.Sum([]byte(text))
	return hash[:]
}
