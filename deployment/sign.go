package deployment

import (
	"crypto/hmac"
	"crypto/sha256"
	"hash"
	"io"
)

func NewHMACCalculator(secret string) hash.Hash {
	return hmac.New(sha256.New, []byte(secret))
}

func CalculateHMAC(rc io.ReadCloser, hmacCalculator hash.Hash) []byte {

	defer rc.Close()
	hmacCalculator.Reset()

	var buf []byte = make([]byte, 16384)
	for n, err := rc.Read(buf); n > 0 && err == nil; n, err = rc.Read(buf) {
		hmacCalculator.Write(buf)
	}

	return hmacCalculator.Sum(nil)
}
