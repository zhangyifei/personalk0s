//go:build linux || darwin
// +build linux darwin

package cmdconfig

import (
	"path/filepath"

	"eke/internal/common"
)

var configPaths = []string{
	"/usr/etc/",
	"/etc/",
	filepath.Join(common.HomeDir(), ".eke"),
}
