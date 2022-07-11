package utils

import (
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"log"
)

func GetIface(name string) (netlink.Link, error) {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		log.Fatalf("cannot find %s: %v", name, err)
		return nil, err
	}
	return iface, nil
}

func CreateFQdisc(iface netlink.Link) (*netlink.Fq, error) {
	//tc qdisc add dev wlp2s0 root fq ce_threshold 4ms
	attrs := netlink.QdiscAttrs{
		LinkIndex: iface.Attrs().Index,
		Handle:    netlink.MakeHandle(0x123, 0),
		Parent:    netlink.HANDLE_ROOT,
	}

	//fq := netlink.NewFq(attrs)

	fq := &netlink.Fq{
		QdiscAttrs: attrs,
		Pacing:     0,
	}

	if err := netlink.QdiscAdd(fq); err != nil {
		log.Fatalf("cannot add fq qdisc: %v", err)
		return nil, err
	}
	log.Printf("Added fq qdisc %v", fq)
	return fq, nil
}

func CreateNetemQdisc(iface netlink.Link, handle, parent uint32, netemAttr netlink.NetemQdiscAttrs) (*netlink.Netem, error) {
	netemQdiscAttr := netlink.QdiscAttrs{
		LinkIndex: iface.Attrs().Index,
		Handle:    handle,
		Parent:    parent,
	}

	netemQdisc := netlink.NewNetem(netemQdiscAttr, netemAttr)

	if err := netlink.QdiscAdd(netemQdisc); err != nil {
		log.Fatalf("cannot add netem qdisc: %v", err)
		return nil, err
	}
	log.Printf("Added netem qdisc %v", netemQdisc)
	return netemQdisc, nil
}

func CreateHtbClass(iface netlink.Link, handle, parent uint32, htbClassAttrs netlink.HtbClassAttrs) (*netlink.HtbClass, error) {
	classAttr := netlink.ClassAttrs{
		LinkIndex: iface.Attrs().Index,
		Handle:    handle,
		Parent:    parent,
	}

	htbClass := netlink.NewHtbClass(classAttr, htbClassAttrs)

	if err := netlink.ClassAdd(htbClass); err != nil {
		log.Fatalf("cannot add htb class: %v", err)
		return nil, err
	}
	log.Printf("Added htb class %v", htbClass)
	return htbClass, nil
}

func CreateHtbQdisc(iface netlink.Link, handle, parent uint32) (*netlink.Htb, error) {
	attrs := netlink.QdiscAttrs{
		LinkIndex: iface.Attrs().Index,
		Handle:    handle,
		Parent:    parent,
	}

	qdiscHtb := netlink.NewHtb(attrs)

	if err := netlink.QdiscAdd(qdiscHtb); err != nil {
		log.Fatalf("cannot add htb qdisc: %v", err)
		return nil, err
	}
	log.Printf("Added htb qdisc %v", qdiscHtb)
	return qdiscHtb, nil
}

func CreateTCBpfFilter(iface netlink.Link, progFd int, parent uint32, name string) (*netlink.BpfFilter, error) {
	filterAttrs := netlink.FilterAttrs{
		LinkIndex: iface.Attrs().Index,
		Parent:    parent,
		Handle:    netlink.MakeHandle(0, 1),
		Protocol:  unix.ETH_P_ALL,
		Priority:  1,
	}

	filter := &netlink.BpfFilter{
		FilterAttrs:  filterAttrs,
		Fd:           progFd,
		Name:         name,
		DirectAction: true,
	}

	if err := netlink.FilterAdd(filter); err != nil {
		log.Fatalf("cannot attach bpf object to filter: %v", err)
		return nil, err
	}
	log.Printf("Created bpf filter: %v", filter)
	return filter, nil
}

func CreateClsactQdisc(iface netlink.Link) (*netlink.GenericQdisc, error) {
	attrs := netlink.QdiscAttrs{
		LinkIndex: iface.Attrs().Index,
		Handle:    netlink.MakeHandle(0xffff, 0),
		Parent:    netlink.HANDLE_CLSACT,
	}

	qdisc := &netlink.GenericQdisc{
		QdiscAttrs: attrs,
		QdiscType:  "clsact",
	}

	if err := netlink.QdiscAdd(qdisc); err != nil {
		log.Fatalf("cannot add clsact qdisc: %v", err)
		return nil, err
	}
	log.Printf("Added clsact qdisc %v", qdisc)
	return qdisc, nil
}
