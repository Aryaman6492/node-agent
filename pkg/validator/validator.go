package validator

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Aryaman6492/node-agent/pkg/config"
	"github.com/Aryaman6492/node-agent/pkg/validator/ebpf"

	"github.com/cilium/ebpf/rlimit"
	"github.com/facette/natsort"
	"github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"golang.org/x/sys/unix"
)

func int8ToStr(arr []int8) string {
	b := make([]byte, 0, len(arr))
	for _, v := range arr {
		if v == 0x00 {
			break
		}
		b = append(b, byte(v))
	}
	return string(b)
}

func checkKernelVersion(minKernelVersion string) error {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return fmt.Errorf("checkKernelVersion: fail to detect the kernel version")
	}
	kernelVersion := int8ToStr(uname.Sysname[:]) + "," + int8ToStr(uname.Release[:]) + "," + int8ToStr(uname.Version[:])
	logger.L().Debug("checkKernelVersion - kernelVersion", helpers.String("is", kernelVersion))

	// use natsort because 5.15 is greater than 5.4 but not from a string comparison perspective
	if natsort.Compare(int8ToStr(uname.Release[:]), minKernelVersion) {
		return fmt.Errorf("checkKernelVersion: the current kernel version %s is less than the min kernel version support %s", int8ToStr(uname.Release[:]), minKernelVersion)
	}

	return nil
}

// see https://github.com/inspektor-gadget/inspektor-gadget/pull/1809
//
// workaroundMounts ensures that filesystems are mounted correctly.
// Some environments (e.g. minikube) runs with a read-only /sys without bpf
// https://github.com/kubernetes/minikube/blob/99a0c91459f17ad8c83c80fc37a9ded41e34370c/deploy/kicbase/entrypoint#L76-L81
// Docker Desktop with WSL2 also has filesystems unmounted.
func workaroundMounts() error {
	fs := []struct {
		name  string
		path  string
		magic int64
	}{
		{
			"bpf",
			"/sys/fs/bpf",
			unix.BPF_FS_MAGIC,
		},
		{
			"debugfs",
			"/sys/kernel/debug",
			unix.DEBUGFS_MAGIC,
		},
		{
			"tracefs",
			"/sys/kernel/tracing",
			unix.TRACEFS_MAGIC,
		},
	}
	for _, f := range fs {
		var statfs unix.Statfs_t
		err := unix.Statfs(f.path, &statfs)
		if err != nil {
			return fmt.Errorf("statfs %s: %w", f.path, err)
		}
		if statfs.Type == f.magic {
			logger.L().Debug("workaroundMounts - already mounted", helpers.String("name", f.name), helpers.String("path", f.path))
		} else {
			err := unix.Mount("none", f.path, f.name, 0, "")
			if err != nil {
				return fmt.Errorf("mounting %s: %w", f.path, err)
			}
			logger.L().Debug("workaroundMounts - mounted", helpers.String("name", f.name), helpers.String("path", f.path))
		}
	}
	return nil
}

func CheckPrerequisites(cfg config.Config) error {
	// Check eBPF support
	logger.L().Debug("CheckPrerequisites - checking eBPF support")
	if err := ebpf.VerifyEbpf(); err != nil {
		return err
	}
	if cfg.KubernetesMode {
		// Check environment variables
		for _, envVar := range []string{config.NodeNameEnvVar, config.PodNameEnvVar, config.NamespaceEnvVar} {
			logger.L().Debug("CheckPrerequisites - checking environment variable", helpers.String("envVar", envVar))
			if getenv := os.Getenv(envVar); getenv == "" {
				return fmt.Errorf("%s environment variable not set", envVar)
			}
		}
	}
	// Ensure all filesystems are mounted
	logger.L().Debug("CheckPrerequisites - checking mounts")
	if err := workaroundMounts(); err != nil {
		return err
	}
	// Raise the rlimit for memlock to the maximum allowed (eBPF needs it)
	logger.L().Debug("CheckPrerequisites - raising memlock rlimit")
	if err := rlimit.RemoveMemlock(); err != nil {
		return err
	}
	return nil
}
