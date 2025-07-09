//go:build linux

package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags linux bpf bpf.c -- -I../headers

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Please specify a process name")
	}

	processName := os.Args[1]

	var dstPort uint16 = 4040
	if len(os.Args) > 2 {
		port, err := strconv.ParseUint(os.Args[2], 10, 16)
		if err != nil {
			log.Fatalf("Invalid port number: %s", err)
		}
		dstPort = uint16(port)
	}

	spec, err := loadBpf()
	if err != nil {
		log.Fatalf("Error loading spec: %s", err)
	}

	objs := bpfObjects{}
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		log.Fatalf("Error loading objects: %s", err)
	}
	defer objs.Close()

	if err := objs.PortMap.Put(uint32(0), dstPort); err != nil {
		log.Fatalf("Failed to update port map: %s", err)
	}

	var proc [16]byte
	copy(proc[:], processName)
	if err := objs.ProcessMap.Put(uint32(0), proc); err != nil {
		log.Fatalf("Failed to update process map: %s", err)
	}

	// Attach LSM hook
	l, err := link.AttachLSM(link.LSMOptions{
		Program: objs.CheckAccept,
	})
	if err != nil {
		log.Fatalf("Failed to attach LSM: %v", err)
	}
	defer l.Close()

	log.Printf("eBPF program loaded. Allowed port: %d for process: %s", dstPort, proc)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Exiting...")
}
