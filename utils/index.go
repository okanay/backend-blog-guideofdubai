package utils

import (
	"crypto/rand"
	"log"
	"math/big"
	"time"
)

const Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// Zaman ölçümü için kullanılır
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	if elapsed <= 5*time.Millisecond {
		return
	}
	log.Printf("%s ~TOOK~ %s", name, elapsed.Round(time.Millisecond))
}

// Güvenli rastgele string üretimi (CSPRNG ile)
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(Alphabet))))
		if err != nil {
			log.Fatalf("crypto/rand failed: %v", err)
		}
		b[i] = Alphabet[num.Int64()]
	}
	return string(b)
}

// Güvenli rastgele sayı üretimi (CSPRNG ile)
func GenerateRandomInt(min, max int) int {
	if max <= min {
		log.Fatalf("Invalid range: max (%d) must be greater than min (%d)", max, min)
	}
	diff := max - min
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		log.Fatalf("crypto/rand failed: %v", err)
	}
	return int(nBig.Int64()) + min
}
