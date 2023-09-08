/*
package cgroup manages the cgroup created by the server
*/
package cgroup

import (
	"errors"
	"log"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup2"
)

const (
	KiB = 2 << 9
	MiB = 2 << 19
	GiB = 2 << 29
)

const (
	USEC = 1                // microsecond
	MSEC = 1_000 * USEC     // millisecond
	SEC  = 1_000_000 * USEC // second
)

type Limit struct {
	Memory int64  // in bytes
	CPU    uint64 // in percent
}

// initializes the cgroup
func Initialize(limit Limit) error {
	if cgroups.Mode() != cgroups.Unified {
		return errors.New("cgroups v2 is not supported")
	}

	period := uint64(100 * MSEC)
	quota := int64(period * limit.CPU / 100)
	res := &cgroup2.Resources{
		Memory: &cgroup2.Memory{
			Max:  &limit.Memory,
			High: &limit.Memory,
		},
		CPU: &cgroup2.CPU{
			Max: cgroup2.NewCPUMax(&quota, &period),
		},
	}

	m, err := cgroup2.NewManager("/sys/fs/cgroup/", "/ddp_playground", res)
	cgroup = m
	return err
}

// destroys the cgroup
func Destroy() error {
	return cgroup.Delete()
}

var cgroup *cgroup2.Manager

func Add(pid uint64) error {
	log.Printf("adding %d to cgroup\n", pid)
	return cgroup.AddProc(pid)
}

func Processes() ([]uint64, error) {
	return cgroup.Procs(true)
}
