package rules

import (
	"fmt"
	"strings"
	"unicode"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/option"
)

const PackageCamelCaseRuleID = "PACKAGE_CAMEL_CASE"

var packageCamelCaseRule = newRule(
	&check.RuleSpec{
		ID:      PackageCamelCaseRuleID,
		Default: false,
		Purpose: "Checks that packages are camelCase: each dot-separated component starts with a lowercase letter and contains no underscores.",
		Type:    check.RuleTypeLint,
	},
	checkPackageCamelCase,
)

func checkPackageCamelCase(file fileSummary, _ option.Options) ([]annotation, error) {
	if file.pkg == "" {
		return nil, nil
	}
	components := strings.Split(file.pkg, ".")
	for i, component := range components {
		components[i] = toCamelCase(component)
	}
	expected := strings.Join(components, ".")
	if expected == file.pkg {
		return nil, nil
	}
	return []annotation{{
		message:    fmt.Sprintf("Package name %q should be camelCase, such as %q.", file.pkg, expected),
		sourcePath: []int32{int32(fileDescriptorProtoPackage.number)},
	}}, nil
}

// toCamelCase joins the words of s in camelCase. Words are separated by
// underscores or by a capital letter that starts a new word, so "HTTP_server",
// "HTTPServer" and "httpServer" all become "httpServer". Uppercase letters
// after the first word are kept, so "userAPI" is already camelCase.
func toCamelCase(s string) string {
	var b strings.Builder
	for i, word := range splitWords(s) {
		runes := []rune(word)
		if i == 0 {
			for j, r := range runes {
				if !unicode.IsUpper(r) {
					break
				}
				runes[j] = unicode.ToLower(r)
			}
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		b.WriteString(string(runes))
	}
	return b.String()
}

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
