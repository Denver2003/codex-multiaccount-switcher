package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func HashNormalized(normalized []byte) string {
	sum := sha256.Sum256(normalized)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:]))
}

func Hash(raw []byte) (string, error) {
	normalized, err := Normalize(raw)
	if err != nil {
		return "", err
	}

	return HashNormalized(normalized), nil
}
