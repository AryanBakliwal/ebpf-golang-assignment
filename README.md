# eBPF Golang Assignment

The `go-server` directory contains a server that listens on two ports - 4040 and 5050.
Start the server by running the following command
```bash
$ ./myprocess # runs a process named myprocess
```

### 1. Drop packets on a port
Uses XDP.

Arguments:

1. Network interface name (required, for example `eth0`, `lo`, etc.)
2. Port (optional, defaults to 4040 if not specified)

Build by running
```bash
$ go generate
$ go build -o packet-port .
```
Commands to run the program
```bash
$ sudo ./packet-port lo # drops packets at port 4040 of lo network interface

$ sudo ./packet-port lo 5050 # drops packets at port 5050 of lo network interface
```

### 2. Allow packets only at a specific port for a process
Uses `lsm/socket_accept`.

Arguments:

1. Process name (required, for example `myprocess`)
2. Port (optional, defaults to 4040 if not specified)

NOTE: Start the eBPF program before starting the server.

Build by running
```bash
$ go generate
$ go build -o packet-process .
```
Commands to run the program
```bash
$ sudo ./packet-process myprocess # allows packets only at port 4040 and drops at all other ports

$ sudo ./packet-process myprocess 5050 # allows packets only at port 5050 and drops at all other ports
```