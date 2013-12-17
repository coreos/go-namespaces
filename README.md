## go-namespaces

Golang implementations of Linux Namespaces

### nspipe

nspipe is a simple example application included with go-namespaces.
It lets you bind a socket into a namespace and connect the other end to some other tcp address.

Outside the namespace in a namespace with routable internet networking:

```sh
nspipe -t $TARGET_PID
```

Inside the namespace with private networking:

```sh
telnet 127.0.0.1 23
```

### Libraries

- [github.com/coreos/go-namespaces/net](http://godoc.org/github.com/coreos/go-namespaces/net)
- [github.com/coreos/go-namespaces/namespaces](http://godoc.org/github.com/coreos/go-namespaces/namespaces)
