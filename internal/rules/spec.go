// Package rules contains the lint rules provided by buf-plugin-lint-extra.
package rules

import (
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
)

var allRules = []rule{
	enumDedicatedFileRule,
	enumFileSuffixRule,
}

var Spec = &check.Spec{
	Rules: ruleSpecs(allRules),
	Info: &info.Spec{
		Documentation: "Extra lint rules for buf: keeps enums in dedicated files and enforces a naming convention for those files.",
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
