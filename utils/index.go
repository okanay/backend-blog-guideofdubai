package utils

import (
	"log"
	"math/rand"
	"time"
)

const Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)

	if elapsed <= 5*time.Millisecond {
		return
	}

	log.Printf("%s ~TOOK~ %s", name, elapsed.Round(time.Millisecond))
}

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = Alphabet[seededRand.Intn(len(Alphabet))]
	}

	return string(b)
}

func GenerateRandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}
