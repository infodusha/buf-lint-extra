package rules

import "unicode"

func splitWords(s string) []string {
	runes := []rune(s)
	var words []string
	start := -1
	for i, r := range runes {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			if start >= 0 {
				words = append(words, string(runes[start:i]))
				start = -1
			}
		case start < 0:
			start = i
		case startsWord(runes, i):
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	if start >= 0 {
		words = append(words, string(runes[start:]))
	}
	return words
}

// startsWord reports whether the capital letter at i begins a word: it follows
// a lowercase letter, as in "userService", or precedes one, as in "HTTPServer".
func startsWord(runes []rune, i int) bool {
	if !unicode.IsUpper(runes[i]) {
		return false
	}
	return unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))
}
