package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// Creates and returns a new hmac object
func NewHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC {
		hmac: h,
	}
}

// A wrapper for the crypto/hmac package to make
// this code cleaner
type HMAC struct {
	hmac hash.Hash
}

func (h HMAC) Hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	bytes := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(bytes)
}