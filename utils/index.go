package utils

import (
	"math/rand"
	"time"
)

const TokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = TokenAlphabet[seededRand.Intn(len(TokenAlphabet))]
	}

	return string(b)
}

func GenerateRandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}
