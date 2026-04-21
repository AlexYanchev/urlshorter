package repository

import (
	"crypto/rand"
	"encoding/hex"
)

func generateUUID() string {
    b := make([]byte, 16)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}