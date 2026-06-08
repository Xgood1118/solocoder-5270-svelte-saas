package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSignature(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifySignature(payload, secret, signature string) bool {
	expected := GenerateSignature(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
