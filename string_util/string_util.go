package stringutil

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

//Sha256HexInt is int to HeshHexString
func Sha256HexInt(src int) string {
	return Sha256HexString(strconv.Itoa(src))
}

//Sha256HexString is string to HeshHexString
func Sha256HexString(src string) string {
	return src
	buf := sha256.Sum256([]byte(src))
	srcByte := make([]byte, 32)
	copy(srcByte[:], buf[:])

	dst := make([]byte, hex.EncodedLen(len(srcByte)))
	hex.Encode(dst, srcByte)

	return string(dst)

}

//Sha256HexString is string to HeshHexString
func Sha256HexString2(src string) string {
	buf := sha256.Sum256([]byte(src))
	srcByte := make([]byte, 32)
	copy(srcByte[:], buf[:])

	dst := make([]byte, hex.EncodedLen(len(srcByte)))
	hex.Encode(dst, srcByte)

	return string(dst)

}
