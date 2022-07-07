//go:build windows
// +build windows

package cmdconfig

import (
	"eke/internal/common"
	"os"
	"path/filepath"
)

var configPaths = []string{
	filepath.Join(os.Getenv("APPDATA"), "eke"),
	filepath.Join(os.Getenv("PROGRAMDATA"), "eke"),
	filepath.Join(common.HomeDir(), ".eke"),
}
