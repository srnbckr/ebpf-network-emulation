package main

import (
	"ebpf-network-simulator/internal/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	bar "github.com/schollz/progressbar"
	"github.com/vishvananda/netlink"
)

const (
	HOSTINTERFACE   = "ens4"
	BRIDGEINTERFACE = "br0"
	GUESTINTERFACE  = "eth0"
	MKTAP           = "./mk-tap.sh"
	CRTAP           = "./create-tap.sh"
	CRBR            = "./create-bridge.sh"
	TC              = "/sbin/tc"
	DEFAULTRATE     = 32000000.0 * 1024 // 32.0 Gbps
)

func main() {
	f, err := os.Create("./setup-log.csv")
	if err != nil {
		panic(1)
	}
	defer f.Close()

	f.WriteString("N,index,i,j,time\n")

	for i := 2 << 6; i > 0; i = i >> 1 {
		log.Printf("starting new process with N = ", i)
		timeNewSync(i, f)
		// timeNewAsync(i, f)
	}

}

func timeNewSync(N int, f *os.File) {
	_, tapDevices := cleanNew(N)

	pbar := bar.New(N * (N - 1))

	start := time.Now()

	// both for-loops now cover an upper triangle (every possible unordered pairing of IPs, excl self)
	for i := 0; i < N; i++ {
		//tapName := "tap" + strconv.Itoa(i)

		for j := 0; j < N; j++ {
			// skip self
			if j == i {
				continue
			}

			//a := getNetwork(uint32(i))
			b := getNetwork(uint32(j))

			innerStart := time.Now()

			//index := uint32(i*N + j)
			index := uint16(j + 2)

			// func CreateNetemForLink(iface netlink.Link, htbRootQdisc *netlink.Htb, index uint16, ip string) error {
			iface, err := netlink.LinkByIndex(tapDevices[fmt.Sprintf("tap%d", i)])
			if err != nil {
				log.Fatalf("Error getting tap interface: %v", err)
			}

			err = createLinkNew(iface, index, b)

			if err != nil {
				log.Fatalf("error: ", err)
			}

			/*err = updateDelayNew(tapMap[tapName], rtnl, index, 1.1, 5)
			if err != nil {
				log.Debugf("error: ", err)
			}*/

			innerEnd := time.Now()
			f.WriteString(fmt.Sprintf("%d,%d,%d,%d,%d\n", N, index, i, j, innerEnd.Sub(innerStart).Nanoseconds()))

			pbar.Add(1)
		}
	}

	elapsed := time.Since(start)

	log.Printf("finished new process in %s", elapsed)
}

func cleanNew(N int) (int, map[string]int) {
	log.Printf("start cleaning and creating tap devices")

	devId, err := createBridge()

	if err != nil {
		log.Fatalf("error: ", err)
	}

	_, err = createQDiscNew(devId)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	tapDevices := make(map[string]int)

	pbar := bar.New(N)
	for i := 0; i < N; i++ {
		tapName := "tap" + strconv.Itoa(i)

		tapDevices[tapName], err = createTap(tapName, fmt.Sprintf("%s/32", getNetwork(uint32(i))))

		if err != nil {
			log.Fatalf("error: ", err)
		}

		_, err = createQDiscNew(tapDevices[tapName])

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		pbar.Add(1)

	}

	log.Printf("finished cleaning and creating tap devices")

	return devId, tapDevices
}

func createBridge() (int, error) {
	log.Printf("creating br0")

	cmd := exec.Command(CRBR)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return -1, errors.Wrapf(err, "%#v: output: %s", cmd.Args, out)
	}

	// generate device id for network interface from script output
	lineSplit := strings.Split(string(out), "\n")
	outSplit := strings.Split(lineSplit[len(lineSplit)-2], ": ")
	devId, err := strconv.Atoi(outSplit[0])

	log.Printf("successfully created br0, device ID %d", devId)

	return devId, nil
}

func createTap(tapName string, addr string) (int, error) {
	log.Printf("creating tap %s", tapName)

	cmd := exec.Command(CRTAP, tapName, BRIDGEINTERFACE, addr)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return 0, errors.Wrapf(err, "%#v: %s", cmd.Args, string(out))
	}

	// generate device id for network interface from script output
	lineSplit := strings.Split(string(out), "\n")
	outSplit := strings.Split(lineSplit[len(lineSplit)-2], ": ")
	devId, err := strconv.Atoi(outSplit[0])

	log.Printf("successfully created tap %s, device ID %d", tapName, devId)

	return devId, nil
}

func removeRootQDisc(allowFail bool) error {
	// tc qdisc del dev [TAP_NAME] root
	cmd := exec.Command(TC, "qdisc", "del", "dev", "br0", "root")

	if out, err := cmd.CombinedOutput(); !allowFail && err != nil {
		return errors.Wrapf(err, "%#v: output: %s", cmd.Args, out)
	}

	return nil
}

func createQDiscNew(devId int) (int64, error) {
	err := removeRootQDisc(true)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	iface, err := netlink.LinkByIndex(devId)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	// create tc htb root qdisc
	htbRootQdisc, err := utils.CreateHtbQdisc(iface, netlink.MakeHandle(1, 0), netlink.HANDLE_ROOT)
	if err != nil {
		log.Fatalf("cannot add htb qdisc: %v", err)
		return 0, err
	}

	// create tc htb class
	classAttr := netlink.HtbClassAttrs{
		Rate:    32000000.0 * 1024,
		Quantum: 1514,
	}
	_, err = utils.CreateHtbClass(iface, netlink.MakeHandle(1, 1), htbRootQdisc.Handle, classAttr)

	if err != nil {
		return 0, errors.WithStack(err)
	}

	// return starting index for delays
	return 1, nil
}

func createLinkNew(iface netlink.Link, index uint16, ip string) error {
	// create htb class
	// (TC, "class", "add", "dev", tapName, "parent", "1:", "classid", fmt.Sprintf("1:%x", index), "htb", "rate", DEFAULTRATE, "quantum", "1514")

	rootHandle := netlink.MakeHandle(1, 0)

	classAttr := netlink.HtbClassAttrs{
		Rate:    DEFAULTRATE,
		Quantum: 1514,
	}
	htbClass, err := utils.CreateHtbClass(iface, netlink.MakeHandle(1, index), rootHandle, classAttr)

	if err != nil {
		log.Fatalf("cannot add htb class: %v", err)
		return err
	}

	cmd := exec.Command(TC, "qdisc", "add", "dev", iface.Attrs().Name, "parent", netlink.HandleStr(htbClass.Handle), "handle", fmt.Sprintf("%x:", index), "netem", "delay", "100.0ms", "limit", "1000000")

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("cannot add netem qdisc: %v  -> %#v   - %s", err, cmd.Args, out)
		return err
	}

	// yes its required to set "src" as "dest_net" and "dst" as "source_net", this is intentional
	// removing the "match ip dst [SOURCE_NET]" filter as it wouldn't do anything: there is only one network on this tap anyway
	// tc filter add dev [TAP_NAME] protocol ip parent 1: prio [INDEX] u32 match ip src [DEST_NET] classid 1:[INDEX]
	cmd = exec.Command(TC, "filter", "add", "dev", iface.Attrs().Name, "protocol", "ip", "parent", netlink.HandleStr(rootHandle), "prio", "1", "u32", "match", "ip", "src", ip, "classid", netlink.HandleStr(htbClass.Handle))

	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("cannot add tc filter: %v  -> %#v   - %s", err, cmd.Args, out)
		return err
	}
	return nil
}

// returns a network in CIDR notation for a tap device with index id
func getNetwork(id uint32) string {
	id = id + 1
	// use a /32 subnet for the tap device
	return fmt.Sprintf("10.%d.%d.%d", (id >> 14 & 0xFF), (id >> 6 & 0xFF), (id << 2 & 0xFF))
}
