//go:build windows
// +build windows

/*
Copyright 2022 eke authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package constant

import "fmt"

const (
	// DataDirDefault folder contains all eke state
	DataDirDefault = "C:\\var\\lib\\eke"
	// CertRootDir defines the root location for all pki related artifacts
	CertRootDir = "C:\\var\\lib\\eke\\pki"
	// BinDir defines the location for all pki related binaries
	BinDir = "C:\\var\\lib\\eke\\bin"
	// RunDir run directory
	RunDir = "C:\\run\\eke"
	// ManifestsDir stack applier directory
	ManifestsDir = "C:\\var\\lib\\eke\\manifests"
	// KubeletVolumePluginDir defines the location for kubelet plugins volume executables
	KubeletVolumePluginDir = "C:\\usr\\libexec\\eke\\kubelet-plugins\\volume\\exec"

	KineSocket                     = "kine\\kine.sock:2379"
	KubePauseContainerImage        = "mcr.microsoft.com/oss/kubernetes/pause"
	KubePauseContainerImageVersion = "1.4.1"
	EkeConfigPathDefault           = "C:\\etc\\eke\\eke.yaml"
)

func formatPath(dir string, file string) string {
	return fmt.Sprintf("%s\\%s", dir, file)
}
