package main

import (
	"bufio"
	"bytes"
	"ebpf-network-simulator/internal/utils"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/cilium/ebpf"
	bar "github.com/schollz/progressbar"
	"github.com/vishvananda/netlink"
)

var (
	iface_name *string
	rules      *int
	test_vm    *string
	run_ebpf   *bool
)

const (
	DEFAULTRATE = 32000000.0 * 1024 // 32.0 Gbps
	TC          = "/sbin/tc"
)

func init() {
	iface_name = flag.String("iface", "eth0", "Interface to attach ebpf program to")
	rules = flag.Int("rules", 10, "Amount of filter rules to add. Refers to qdiscs/filters or eBPF map entries")
	test_vm = flag.String("test_host", "172.18.0.2:8888", "Test Unikernel VM")
	run_ebpf = flag.Bool("run_ebpf", false, "Run the eBPF experiment")
}

func main() {
	flag.Parse()

	iface, err := utils.GetIface(*iface_name)
	if err != nil {
		log.Fatalf("cannot find %s: %v", *iface_name, err)
	}

	if *run_ebpf {
		ebpfExp(iface)
		return
	}
	classic(iface)
}

type handleBpsDelay struct {
	tcHandle        uint32
	throttleRateBps uint32
	delayMs         uint32
}

func parseStringToLong(ip string) uint32 { // TODO: add error handling
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.LittleEndian, &long)
	return long
}

// Expect that an ebpf program was attached to the given interface with ebpf-network-simulation
func ebpfExp(iface netlink.Link) {
	// Path to the map file of the eBPF program
	ebpfMapFile := "/sys/fs/bpf/IP_HANDLE_BPS_DELAY"

	// Load map
	ebpfMap, err := ebpf.LoadPinnedMap(ebpfMapFile, &ebpf.LoadPinOptions{})
	if err != nil {
		fmt.Println("err: something went wrong with loading the pinned map")
		fmt.Println(err)
		os.Exit(1)
	}

	N := *rules // fill map with N entries

	pbar := bar.New(*rules)

	c, err := net.Dial("tcp", *test_vm)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	nextWait := 1
	start := time.Now()

	log.Printf("Start with 0 filters. Waiting")
	fmt.Fprintf(c, strconv.Itoa(0)+"\n")
	_, err = bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for i := 0; i < N; i++ {
		var handle_bps_delay handleBpsDelay
		index := int64(i + 2)
		ip_string := fmt.Sprintf("172.16.%d.%d", index/256, index%256)

		// set the same values for all entries for now
		handle_bps_delay.tcHandle = uint32(1)
		handle_bps_delay.throttleRateBps = 5000000
		handle_bps_delay.delayMs = uint32(i + 10)

		err := ebpfMap.Put(parseStringToLong(ip_string), handle_bps_delay)
		if err != nil {
			fmt.Println("err: putting ip and handle into map failed")
			fmt.Println(err)
			os.Exit(1)
		}
		pbar.Add(1)
		if (i+1) == nextWait || (i+1) >= N {
			log.Printf("Created %d map entries. Waiting.", i+1)
			fmt.Fprintf(c, strconv.Itoa(i+1)+"\n")

			if nextWait == 1 {
				nextWait = 500
			} else {
				nextWait = nextWait + 500
			}

			_, err = bufio.NewReader(c).ReadString('\n')
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("\n\nTime elapsed for %d rules: %s\n", N, elapsed)
}

func classic(iface netlink.Link) {
	// create htb root qdisc
	htbRootQdisc, err := utils.CreateHtbQdisc(iface, netlink.MakeHandle(1, 0), netlink.HANDLE_ROOT)
	if err != nil {
		log.Fatalf("cannot add htb qdisc: %v", err)
		return
	}

	N := *rules
	pbar := bar.New(*rules)

	c, err := net.Dial("tcp", *test_vm)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	nextWait := 1

	start := time.Now()

	log.Printf("Start with 0 filters. Waiting")
	fmt.Fprintf(c, strconv.Itoa(0)+"\n")
	_, err = bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for i := 0; i < N; i++ {
		index := int64(i + 2)
		err := createNetemForLink(iface, htbRootQdisc, uint16(index), fmt.Sprintf("172.16.%d.%d", index/256, index%256))
		if err != nil {
			log.Fatalf("error creating netem link: %v", err)
		}
		pbar.Add(1)
		if (i+1) == nextWait || (i+1) >= N {
			log.Printf("Created %d qdiscs and filters. Waiting.", i+1)
			fmt.Fprintf(c, strconv.Itoa(i+1)+"\n")

			if nextWait == 1 {
				nextWait = 500
			} else {
				nextWait = nextWait + 500
			}

			_, err = bufio.NewReader(c).ReadString('\n')
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		}
	}
	elapsed := time.Since(start)
	log.Printf("\nDuration: %v\n", elapsed)
	log.Printf("Cleaning up")
	netlink.QdiscDel(htbRootQdisc)

}

func createNetemForLink(iface netlink.Link, htbRootQdisc *netlink.Htb, index uint16, ip string) error {
	// create htb class
	// (TC, "class", "add", "dev", tapName, "parent", "1:", "classid", fmt.Sprintf("1:%x", index), "htb", "rate", DEFAULTRATE, "quantum", "1514")

	classAttr := netlink.HtbClassAttrs{
		Rate:    DEFAULTRATE,
		Quantum: 1514,
	}
	htbClass, err := utils.CreateHtbClass(iface, netlink.MakeHandle(1, index), htbRootQdisc.Handle, classAttr)

	if err != nil {
		log.Fatalf("cannot add htb class: %v", err)
		return err
	}
	// create netem qdisc
	// (TC, "qdisc", "add", "dev", tapName, "parent", fmt.Sprintf("1:%x", index), "handle", fmt.Sprintf("%x:", index), "netem", "delay", "0.0", "limit", "1000000")
	//netemAttr := netlink.NetemQdiscAttrs{
	//	Latency: 100 * 1000, // 100ms
	//	Limit:   1000000,
	//}
	//_, err = utils.CreateNetemQdisc(iface, parseStringToLong(fmt.Sprintf("%x:", index)), htbClass.Handle, netemAttr)
	//
	//if err != nil {
	//	log.Fatalf("cannot add netem qdisc: %v", err)
	//	return err
	//}

	cmd := exec.Command(TC, "qdisc", "add", "dev", iface.Attrs().Name, "parent", netlink.HandleStr(htbClass.Handle), "handle", fmt.Sprintf("%x:", index), "netem", "delay", "100.0ms", "limit", "1000000")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("cannot add netem qdisc: %v  -> %#v   - %s", err, cmd.Args, out)
		return err
	}

	// create tc filter
	// vishvananda/netlink in v.1.1.0 does not yet contain u32 filters
	//cmd := exec.Command(TC, "filter", "add", "dev", iface.Attrs().Name, "protocol", "ip", "parent", netlink.HandleStr(htbRootQdisc.Handle), "prio", "1", "u32", "match", "ip", "dst", "1.1.1.1/32", "match", "ip", "src", "0.0.0.0/0", "flowid", netlink.HandleStr(htbClass.Handle))

	// yes its required to set "src" as "dest_net" and "dst" as "source_net", this is intentional
	// removing the "match ip dst [SOURCE_NET]" filter as it wouldn't do anything: there is only one network on this tap anyway
	// tc filter add dev [TAP_NAME] protocol ip parent 1: prio [INDEX] u32 match ip src [DEST_NET] classid 1:[INDEX]
	cmd = exec.Command(TC, "filter", "add", "dev", iface.Attrs().Name, "protocol", "ip", "parent", netlink.HandleStr(htbRootQdisc.Handle), "prio", "1", "u32", "match", "ip", "src", ip, "classid", netlink.HandleStr(htbClass.Handle))

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("cannot add tc filter: %v  -> %#v   - %s", err, cmd.Args, out)
		return err
	}
	return nil
}
