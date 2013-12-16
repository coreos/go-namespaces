package net

import (
	gnet "net"
	"os"
	"syscall"

	"github.com/philips/go-namespace/namespace"
)

func socketat(fd, domain, typ, proto int) (gnet.Listener, error) {
	origNs, _ := namespace.OpenNamespace(namespace.CLONE_NEWNET, os.Getpid())
	defer syscall.Close(int(origNs))
	defer namespace.Setns(origNs, namespace.CLONE_NEWNET)

	// Join the container namespace
	err := namespace.Setns(uintptr(fd), namespace.CLONE_NEWNET)
	if err != 0 {
		return nil, err
	}

	// Create our socket
	return gnet.Listen("tcp", ":12345")
}

// ListenNamespace creates a net.Listener in the namespace of the given pid.
func ListenNamespace(pid uintptr) (gnet.Listener, error) {
	fd, err := namespace.OpenNamespace(namespace.CLONE_NEWNET, int(pid))
	defer syscall.Close(int(fd))
	if err != nil {
		return nil, err
	}

	return socketat(int(fd), syscall.AF_INET, syscall.SOCK_STREAM, 0)
}
