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
