package main

import (
	"ebpf-network-simulator/pkg/utils"
	"flag"
	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
	"log"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go tc ebpf/tc_ebpf_delay.c -- -I../headers

const PIN_PATH = "/sys/fs/bpf/"

var (
	iface_name *string
)

func init() {
	iface_name = flag.String("iface", "eth0", "Interface to attach ebpf program to")
}

func main() {
	flag.Parse()
	objs := tcObjects{}

	opts := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: PIN_PATH,
		},
	}

	if err := loadTcObjects(&objs, &opts); err != nil {
		log.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	progFd := objs.tcPrograms.TcMain.FD()
	iface, err := utils.GetIface(*iface_name)
	if err != nil {
		log.Fatalf("cannot find %s: %v", iface_name, err)
	}

	// Create clsact qdisc
	if _, err := utils.CreateClsactQdisc(iface); err != nil {
		log.Fatalf("cannot create clsact qdisc: %v", err)
	}

	// Create direct-action filter
	if _, err := utils.CreateTCBpfFilter(iface, progFd, netlink.HANDLE_MIN_EGRESS, "ebpf_tc_delay"); err != nil {
		log.Fatalf("cannot create bpf filter: %v", err)
	}

	// Create root htb qdisc
	rootHtbQdisc, err := utils.CreateHtbQdisc(iface, netlink.MakeHandle(0x1a1e, 0), netlink.HANDLE_ROOT)
	if err != nil {
		log.Fatalf("cannot create htb qdisc: %v", err)
	}

	// Create htb class
	htbClassAttrs := netlink.HtbClassAttrs{
		Rate:    32000000.0 * 1024,
		Ceil:    0,
		Buffer:  0,
		Cbuffer: 0,
		Quantum: 0,
		Level:   0,
	}

	if _, err := utils.CreateHtbClass(iface, netlink.MakeHandle(0x1a1e, 1), rootHtbQdisc.Attrs().Handle,
		htbClassAttrs); err != nil {
		log.Fatalf("cannot create htb class: %v", err)
	}

	// Create second htb qdisc
	htbClassAttrs2 := netlink.HtbClassAttrs{
		Rate: 32000000 * 1024,
		Ceil: 32000000 * 1024,
	}

	htbLeafClass, err := utils.CreateHtbClass(iface, netlink.MakeHandle(0x1a1e, 3), rootHtbQdisc.Attrs().Handle, htbClassAttrs2)
	if err != nil {
		log.Fatalf("cannot create second htb class: %v", err)
	}

	// Create netem qdisc
	netemAttr := netlink.NetemQdiscAttrs{
		Latency: 200 * 1000,
	}

	_, err = utils.CreateNetemQdisc(iface, netlink.MakeHandle(0x205d, 0), htbLeafClass.Attrs().Handle, netemAttr)
	if err != nil {
		log.Fatalf("cannot create netem qdisc: %v", err)
	}
}
