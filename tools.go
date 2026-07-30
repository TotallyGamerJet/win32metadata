//go:build tools
// +build tools

package win32metadata

import (
	// Pin dependencies of the code generators.
	// Generator sources are build-ignored, so their imports are invisible to go mod tidy.
	_ "golang.org/x/tools/go/packages"
)
