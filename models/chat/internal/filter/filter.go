package filter

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var regexLeadCloseWhitepace = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
var regexInsideWhitespace = regexp.MustCompile(`[\s\p{Zs}]{2,}`)

func FilteredMessage(text string, filteredWords []string) bool {
	if len(filteredWords) == 0 {
		return false
	}

	text = norm.NFC.String(strings.ToLower(text))
	text = strings.ReplaceAll(text, "\t", "")
	text = strings.ReplaceAll(text, "\u200b", "")
	text = regexLeadCloseWhitepace.ReplaceAllString(text, "")
	text = regexInsideWhitespace.ReplaceAllString(text, "")

	textNoSpaces := strings.ReplaceAll(text, " ", "")

	for _, word := range filteredWords {
		if strings.Contains(text, word) {
			return true
		}

		if strings.Contains(textNoSpaces, word) {
			return true
		}
	}
	return false
}
