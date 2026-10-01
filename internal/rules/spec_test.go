package rules

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestSpec(t *testing.T) {
	t.Parallel()
	checktest.SpecTest(t, Spec)
}
