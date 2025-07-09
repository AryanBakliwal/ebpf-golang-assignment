//go:build linux

package main

import (
	"log"
	"net"
	"os"
	"strconv"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags linux bpf bpf.c -- -I../headers

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Please specify a network interface")
	}

	// Look up the network interface by name.
	ifaceName := os.Args[1]
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("Lookup network interface %q failed: %s", ifaceName, err)
	}

	var dstPort uint16 = 4040
	if len(os.Args) > 2 {
		port, err := strconv.ParseUint(os.Args[2], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port number: %s", err)
		}
		dstPort = uint16(port)
	}

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("Error loading objects: %s", err)
	}
	defer objs.Close()

	key := uint32(0)
	err = objs.TargetPortMap.Update(&key, &dstPort, ebpf.UpdateAny)
	if err != nil {
		log.Fatalf("Failed to update target port in BPF map: %s", err)
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProgFunc,
		Interface: iface.Index,
	})
	if err != nil {
		log.Fatalf("Could not attach XDP program: %s", err)
	}
	defer l.Close()

	log.Printf("Attached XDP program to iface %q (index %d)", iface.Name, iface.Index)
	log.Printf("Dropping packets on port %d", dstPort)

	select {}
}
