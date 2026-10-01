// Package main implements buf-plugin-lint-extra, a buf check plugin with extra lint rules.
package main

import (
	"buf.build/go/bufplugin/check"

	"github.com/infodusha/buf-lint-extra/internal/rules"
)

func main() {
	check.Main(rules.Spec)
}
