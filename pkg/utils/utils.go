package utils

import (
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

func ToSnakeCase(s string) string {
	var res strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			res.WriteRune('_')
		}
		res.WriteRune(unicode.ToLower(r))
	}

	return res.String()
}
