// namespace implements low level APIs for moving a process into a given Linux Namespace.
package namespace

import (
	"errors"
	"os"
	"path"
	"strconv"
	"syscall"
)

type Namespace struct {
	Path string
	Type uintptr
}

// Namespaces
const (
	CLONE_NEWNS   = 0x00020000
	CLONE_NEWUTS  = 0x04000000
	CLONE_NEWIPC  = 0x08000000
	CLONE_NEWNET  = 0x40000000
	CLONE_NEWUSER = 0x10000000
	CLONE_NEWPID  = 0x20000000
)

var (
	Types []Namespace
)

func init() {
	Types = []Namespace{
		Namespace{Path: "ns/user", Type: CLONE_NEWUSER},
		Namespace{Path: "ns/ipc", Type: CLONE_NEWIPC},
		Namespace{Path: "ns/uts", Type: CLONE_NEWUTS},
		Namespace{Path: "ns/net", Type: CLONE_NEWNET},
		Namespace{Path: "ns/pid", Type: CLONE_NEWPID},
		Namespace{Path: "ns/mnt", Type: CLONE_NEWNS},
	}
}

// Setns is a wrapper around Syscall for the SYS_SETNS
func Setns(fd uintptr, nstype uintptr) syscall.Errno {
	// TODO: make this work on non-amd64 architectures
	_, _, err := syscall.Syscall(SYS_SETNS, uintptr(fd), uintptr(nstype), 0)
	return err
}

// OpenNamespace opens a file descriptor for a given pid and type and returns
// the open fd. The caller is responsible for closing the fd.
func OpenNamespace(nstype uintptr, pid int) (uintptr, error) {
	var nsPath string

	for _, n := range Types {
		if n.Type == nstype {
			nsPath = path.Join("/", "proc", strconv.Itoa(pid), n.Path)
		}
	}

	if nsPath == "" {
		return 0, errors.New("Cannot find namespace type")
	}

	file, err := os.Open(nsPath)
	if err != nil {
		return 0, err
	}

	return file.Fd(), nil
}
