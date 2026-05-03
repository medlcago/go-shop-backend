package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
	"unicode"

	"github.com/gosimple/slug"
)

func Slugify(text string) string {
	return slug.Make(text)
}

func BuildSearchQuery(input string) string {
	words := strings.Fields(input)

	var parts []string

	for _, word := range words {
		cleaned := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return r
			}
			return -1
		}, word)

		if cleaned != "" {
			parts = append(parts, cleaned+":*")
		}
	}

	return strings.Join(parts, " & ")
}

func GenerateCode(n int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, n)

	for i := 0; i < n; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}

	return string(code), nil
}
