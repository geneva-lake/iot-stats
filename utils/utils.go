package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func GenerateHash(password string) string {
	var hash string
	sha := sha256.New()
	sha.Write([]byte(password))
	hash = base64.URLEncoding.EncodeToString(sha.Sum(nil))
	return hash
}
