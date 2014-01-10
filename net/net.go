// net implements a simple interface to the network namespaces.
package net

import (
	gnet "net"
	"os"
	"syscall"

	"github.com/coreos/go-namespaces/namespace"
)

func socketat(fd int, net, laddr string) (gnet.Listener, error) {
	origNs, _ := namespace.OpenProcess(os.Getpid(), namespace.CLONE_NEWNET)
	defer syscall.Close(int(origNs))
	defer namespace.Setns(origNs, namespace.CLONE_NEWNET)

	// Join the container namespace
	err := namespace.Setns(uintptr(fd), namespace.CLONE_NEWNET)
	if err != 0 {
		return nil, err
	}

	// Create our socket
	return gnet.Listen(net, laddr)
}

// ListenProcessNamespace creates a net.Listener in the namespace of the given pid.
// The arguments are identical to net.Listen.
func ListenProcessNamespace(pid uintptr, net, laddr string) (gnet.Listener, error) {
	fd, err := namespace.OpenProcess(int(pid), namespace.CLONE_NEWNET)
	defer namespace.Close(fd)
	if err != nil {
		return nil, err
	}

	return socketat(int(fd), net, laddr)
}

// ListenNamespace creates a net.Listener in the namespace of the given namespace path.
// The arguments are identical to net.Listen.
func ListenNamespace(nsPath string, net, laddr string) (gnet.Listener, error) {
	fd, err := namespace.Open(nsPath)
	defer namespace.Close(fd)
	if err != nil {
		return nil, err
	}

	return socketat(int(fd), net, laddr)
}
