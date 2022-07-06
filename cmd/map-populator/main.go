package main

import (
	"flag"
	"fmt"
	"github.com/cilium/ebpf"
	bar "github.com/schollz/progressbar"
	"os"
	"strconv"
	"strings"
	"time"
)

// HANDLE_BPS map value struct
type handle_bps_delay struct {
	tcHandle        uint32
	throttleRateBps uint32
	delayMs         uint32
}

// parseIp parses IP address from string into uint32 (with reversed order)
func parseIp(ip string) uint32 { // TODO: add error handling
	// split ip address
	split := strings.Split(ip, ".")
	ipParts := make([]uint32, 4)

	// parse into uint32
	for i, s := range split {
		num, err := strconv.Atoi(s)

		if err != nil {
			fmt.Println("error: " + err.Error())
		}

		ipParts[i] = uint32(num)
	}

	// build into one number
	return ((ipParts[3] << 24) & 0xff000000) | ((ipParts[2] << 16) & 0x00ff0000) | ((ipParts[1] << 8) & 0x0000ff00) | (ipParts[0] & 0x000000ff)
}

func fillMap(ebpfMap *ebpf.Map) {
	N := 65534 // fill map with N entries

	pbar := bar.New(N)
	start := time.Now()

	for i := 0; i < N; i++ {
		var handle_bps_delay handle_bps_delay
		index := int64(i + 2)
		ip_string := fmt.Sprintf("172.16.%d.%d", index/256, index%256)

		// set the same values for all entries for now
		handle_bps_delay.tcHandle = uint32(1)
		handle_bps_delay.throttleRateBps = 5000000
		handle_bps_delay.delayMs = 100

		err := ebpfMap.Put(parseIp(ip_string), handle_bps_delay)
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
	var ip uint32 = 0x01010101     // IP 1.1.1.1
	var handle uint32 = 0x1a1e0003 // tc handle 1a1e:3
	var handleBpsMapValue handle_bps_delay

	handleBpsMapValue.tcHandle = handle
	handleBpsMapValue.throttleRateBps = 5000000
	handleBpsMapValue.delayMs = 100

	err = ipHandleMap.Put(ip, handleBpsMapValue)
	if err != nil {
		fmt.Println("err: putting ip and handle into map failed")
		fmt.Println(err)
		os.Exit(1)
	}
}
