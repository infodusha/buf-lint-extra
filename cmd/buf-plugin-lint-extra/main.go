// Package main implements buf-plugin-lint-extra, a buf check plugin with extra lint rules.
package main

import (
	"runtime/debug"

	"pluginrpc.com/pluginrpc"

	"github.com/infodusha/buf-lint-extra/internal/rules"
)

func main() {
	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(512 << 20)
	pluginrpc.Main(rules.NewServer)
}
