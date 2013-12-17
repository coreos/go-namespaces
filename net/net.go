// net implements a simple interface to the network namespaces.
package net

import (
	gnet "net"
	"os"
	"syscall"

	"github.com/coreos/go-namespaces/namespaces"
)

func socketat(fd int, net, laddr string) (gnet.Listener, error) {
	origNs, _ := namespace.OpenNamespace(namespace.CLONE_NEWNET, os.Getpid())
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

// ListenNamespace creates a net.Listener in the namespace of the given pid.
// The arguments are identical to net.Listen.
func ListenNamespace(pid uintptr, net, laddr string) (gnet.Listener, error) {
	fd, err := namespace.OpenNamespace(namespace.CLONE_NEWNET, int(pid))
	defer syscall.Close(int(fd))
	if err != nil {
		return nil, err
	}

	return socketat(int(fd), net, laddr)
}
