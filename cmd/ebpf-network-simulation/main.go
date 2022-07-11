package main

import (
	"ebpf-network-simulator/internal/utils"
	"flag"
	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
	"log"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go edt ebpf/network_simulation.c -- -I../headers

const (
	PIN_PATH = "/sys/fs/bpf/"
)

var (
	iface_name *string
)

func init() {
	iface_name = flag.String("iface", "eth0", "Interface to attach ebpf program to")
}

func main() {
	flag.Parse()
	objs := edtObjects{}

	opts := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: PIN_PATH,
		},
	}

	if err := loadEdtObjects(&objs, &opts); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	progFd := objs.edtPrograms.TcMain.FD()
	iface, err := utils.GetIface(*iface_name)
	if err != nil {
		log.Fatalf("cannot find %s: %v", iface_name, err)
	}

	// Create clsact qdisc
	if _, err := utils.CreateClsactQdisc(iface); err != nil {
		log.Fatalf("cannot create clsact qdisc: %v", err)
	}

	// Create fq qdisc
	if _, err := utils.CreateFQdisc(iface); err != nil {
		log.Fatalf("cannot create fq qdisc: %v", err)
	}

	// Attach bpf program
	if _, err := utils.CreateTCBpfFilter(iface, progFd, netlink.HANDLE_MIN_EGRESS, "edt_bandwidth"); err != nil {
		log.Fatalf("cannot create bpf filter: %v", err)
	}

	// Update jump map with delay prog
	err = objs.Progs.Update(uint32(0), uint32(objs.SetDelay.FD()), ebpf.UpdateAny)
	if err != nil {
		println("Update", err.Error())
	}
}
