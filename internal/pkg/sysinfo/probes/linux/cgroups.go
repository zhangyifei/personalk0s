//go:build linux
// +build linux

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

package linux

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"

	"eke/internal/pkg/sysinfo/probes"
)

type CgroupsProbes struct {
	path              probes.ProbePath
	probeUname        unameProber
	probeCgroupSystem cgroupSystemProber
	probes            probes.Probes
}

func (p *LinuxProbes) RequireCgroups() *CgroupsProbes {
	var c *CgroupsProbes
	p.probes.Set("cgroups", func(path probes.ProbePath, current probes.Probe) probes.Probe {
		if probe, ok := current.(*CgroupsProbes); ok {
			c = probe
			return c
		}

		c = newCgroupsProbes(append(p.path, path...), p.probeUname, "/sys/fs/cgroup")
		return c
	})

	return c
}

func newCgroupsProbes(path probes.ProbePath, unameProber unameProber, mountPoint string) *CgroupsProbes {
	return &CgroupsProbes{
		path,
		unameProber,
		newCgroupSystemProber(unameProber, mountPoint),
		probes.NewProbes(),
	}
}

func (c *CgroupsProbes) Path() probes.ProbePath {
	return c.path
}

func (c *CgroupsProbes) DisplayName() string {
	return "Control Groups"
}

func (c *CgroupsProbes) Probe(reporter probes.Reporter) error {
	if err := c.probeSystem(reporter); err != nil {
		return err
	}

	return c.probes.Probe(reporter)
}

func (c *CgroupsProbes) probeSystem(reporter probes.Reporter) error {
	sys, err := c.probeCgroupSystem()
	if err != nil {
		return reportCgroupSystemErr(reporter, c, err)
	}
	return reporter.Pass(c, sys)
}

func reportCgroupSystemErr(reporter probes.Reporter, d probes.ProbeDesc, err error) error {
	if detectionFailed, ok := err.(cgroupFsDetectionFailed); ok {
		return reporter.Reject(d, detectionFailed, "")
	}

	return reporter.Error(d, err)
}

type cgroupSystem interface {
	probes.ProbedProp
	probeController(string) (cgroupControllerAvailable, error)
	loadControllers(func(name, msg string)) error
}

type cgroupSystemProber func() (cgroupSystem, error)

func loadCgroupSystem(probeUname unameProber, mountPoint string) (cgroupSystem, error) {
	// https://man7.org/linux/man-pages/man7/cgroups.7.html

	var st syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &st); err != nil {
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("no file system mounted at %q", mountPoint)
			return nil, cgroupFsDetectionFailed(msg)
		}

		return nil, fmt.Errorf("failed to stat %q: %w", mountPoint, err)
	}

	switch st.Type {
	case unix.CGROUP2_SUPER_MAGIC:
		// https://www.kernel.org/doc/html/v5.16/admin-guide/cgroup-v2.html#mounting
		return &cgroupV2{probeUname: probeUname}, nil
	case unix.CGROUP_SUPER_MAGIC, unix.TMPFS_MAGIC:
		// https://git.kernel.org/pub/scm/docs/man-pages/man-pages.git/tree/man7/cgroups.7?h=man-pages-5.13#n159
		// https://www.kernel.org/doc/html/v5.16/admin-guide/cgroup-v1/cgroups.html#how-do-i-use-cgroups
		return &cgroupV1{}, nil
	default:
		msg := fmt.Sprintf("unexpected file system type of %q: 0x%x", mountPoint, st.Type)
		return nil, cgroupFsDetectionFailed(msg)
	}
}

type cgroupFsDetectionFailed string

func (c cgroupFsDetectionFailed) Error() string {
	return string(c)
}

func (c cgroupFsDetectionFailed) String() string {
	return string(c)
}

func newCgroupSystemProber(probeUname unameProber, mountPoint string) cgroupSystemProber {
	var once sync.Once
	var sys cgroupSystem
	var err error

	return func() (cgroupSystem, error) {
		once.Do(func() {
			sys, err = loadCgroupSystem(probeUname, mountPoint)
		})

		return sys, err
	}
}
