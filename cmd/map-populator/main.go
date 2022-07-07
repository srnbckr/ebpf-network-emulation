package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cilium/ebpf"
	bar "github.com/schollz/progressbar"
)

// HANDLE_BPS map value struct
type handleBpsDelay struct {
	tcHandle        uint32
	throttleRateBps uint32
	delayMs         uint32
}

// parseIp parses IP address from string into uint32 (with reversed order)
func parseIpToLong(ip string) uint32 { // TODO: add error handling
	var long uint32
	binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.LittleEndian, &long)
	return long
}

func fillMap(ebpfMap *ebpf.Map) {
	N := 65534 // fill map with N entries

	pbar := bar.New(N)
	start := time.Now()

	for i := 0; i < N; i++ {
		var handle_bps_delay handleBpsDelay
		index := int64(i + 2)
		ip_string := fmt.Sprintf("172.16.%d.%d", index/256, index%256)

		// set the same values for all entries for now
		handle_bps_delay.tcHandle = uint32(1)
		handle_bps_delay.throttleRateBps = 5000000
		handle_bps_delay.delayMs = uint32(i + 10)

		err := ebpfMap.Put(parseIpToLong(ip_string), handle_bps_delay)
		if err != nil {
			fmt.Println("err: putting ip and handle into map failed")
			fmt.Println(err)
			os.Exit(1)
		}
		pbar.Add(1)
	}
	elapsed := time.Since(start)
	fmt.Printf("\n\nTime elapsed for %d rules: %s\n", N, elapsed)
}

func main() {
	var unpinMapMode bool

	flag.BoolVar(&unpinMapMode, "unpin-map", false, "Unpins the map and exits.")

	flag.Parse()

	// Path to the map file of the eBPF program
	ebpfMapFile := "/sys/fs/bpf/IP_HANDLE_BPS_DELAY"

	// Load map
	ipHandleMap, err := ebpf.LoadPinnedMap(ebpfMapFile, &ebpf.LoadPinOptions{})
	if err != nil {
		fmt.Println("err: something went wrong with loading the pinned map")
		fmt.Println(err)
		os.Exit(1)
	}

	// Check if map should be unpinned
	if unpinMapMode {
		err = ipHandleMap.Unpin()
		if err != nil {
			fmt.Println("err: could not unpin map")
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Print map
	fmt.Printf("Loaded Map: %+v\n", ipHandleMap)

	// fill the map
	fillMap(ipHandleMap)

	// Add another entry to the map to test overhead
	var ip uint32 = parseIpToLong("1.1.1.1") // IP 1.1.1.1
	//ip_string := fmt.Sprintf("46.4.%d.%d", 61, 148)
	var handle uint32 = 0x1a1e0003 // tc handle 1a1e:3
	var handleBpsMapValue handleBpsDelay

	handleBpsMapValue.tcHandle = handle
	handleBpsMapValue.throttleRateBps = 10000000
	handleBpsMapValue.delayMs = 100

	err = ipHandleMap.Put(ip, handleBpsMapValue)
	if err != nil {
		fmt.Println("err: putting ip and handle into map failed")
		fmt.Println(err)
		os.Exit(1)
	}
}
