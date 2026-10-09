// Package rules contains the lint rules provided by buf-plugin-lint-extra.
package rules

import (
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
)

var allRules = []rule{
	directorySamePackageExtraRule,
	enumDedicatedFileRule,
	enumDedicatedPackageRule,
	enumFileMatchRule,
	enumFileSuffixRule,
	fileLowerKebabCaseRule,
	packageCamelCaseRule,
	packageDirectoryMatchExtraRule,
}

var Spec = &check.Spec{
	Rules: ruleSpecs(allRules),
	Info: &info.Spec{
		Documentation: "Extra lint rules for buf: keeps enums in dedicated files and packages named after them, checks naming conventions for enum files, file names, and packages, matches packages to directories with configurable prefixes and case, and keeps the files of a directory in one package with enum packages counted as their parent.",
		SPDXLicenseID: "Apache-2.0",
		LicenseURL:    "https://github.com/infodusha/buf-lint-extra/blob/main/LICENSE",
	},
}

func ruleSpecs(rules []rule) []*check.RuleSpec {
	specs := make([]*check.RuleSpec, len(rules))
	for i, rule := range rules {
		specs[i] = rule.spec
	}
	return specs
}
