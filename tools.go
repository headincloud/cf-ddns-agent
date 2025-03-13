//go:build tools
// +build tools

package tools

import (
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/mgechev/revive"
	_ "golang.org/x/tools/cmd/goimports"
)
