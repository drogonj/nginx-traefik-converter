package util

import (
	"crypto/sha1"
	"encoding/hex"
)

func ShortHash(input string) string {
	sum := sha1.Sum([]byte(input))
	return hex.EncodeToString(sum[:])[:8]
}
