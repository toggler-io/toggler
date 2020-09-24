package security

import (
	"crypto/sha512"
	"encoding/hex"
)

func ToSHA512Hex(text string) (string, error) {
	hasher := sha512.New()

	if _, err := hasher.Write([]byte(text)); err != nil {
		return "", nil
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
