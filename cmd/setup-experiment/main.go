package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/florianl/go-tc"
	"github.com/pkg/errors"
	bar "github.com/schollz/progressbar"
	log "github.com/sirupsen/logrus"
)

const (
	DEFAULTRATE = "10.0Gbps"

	// MAXLATENCY means unusable: nothing is usable above 999.999 seconds?
	MAXLATENCY   = 999999.9
	MINBANDWIDTH = 0

	HOSTINTERFACE   = "ens4"
	BRIDGEINTERFACE = "br0"
	GUESTINTERFACE  = "eth0"
	NAMESERVER      = "1.1.1.1"

	WGPORT      = 3000
	WGINTERFACE = "wg0"
	MASK        = "/26"

	TC       = "/sbin/tc"
	IPTABLES = "/sbin/iptables"
	IPSET    = "/sbin/ipset"
	MKTAP    = "./mk-tap.sh"
	CRTAP    = "./create-tap.sh"
	CRBR     = "./create-bridge.sh"
)

func main() {
	// f, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_CREATE, 0755)
	// if err != nil {
	// 	panic(1)
	// }
	// log.SetOutput(f)

	f, err := os.Create("./log.csv")
	if err != nil {
		panic(1)
	}
	defer f.Close()

	f.WriteString("N,index,i,j,time\n")

	log.SetOutput(os.Stdout)
	// log.SetLevel(log.DebugLevel)

	// for i := 1; i < 2<<12; i = i << 1 {
	for i := 2 << 8; i > 0; i = i >> 1 {
		log.Info("starting new process with N = ", i)
		timeNewSync(i, f)
		// timeNewAsync(i, f)
	}

	// N := 500

	// timeNewSync(N)

	//timeNewAsync(N)

}

func cleanNew(N int, rtnl *tc.Tc) (int, map[string]int) {
	log.Debug("start cleaning and creating tap devices")

	devId, err := createBridge()

	if err != nil {
		log.Debug("error: ", err)
	}

	_, err = createQDiscNew(devId, rtnl)

	if err != nil {
		log.Debugf("error: %v", err)
	}

	tapDevices := make(map[string]int)

	pbar := bar.New(N)
	for i := 0; i < N; i++ {
		tapName := "tap" + strconv.Itoa(i)

		tapDevices[tapName], err = createTap(tapName, fmt.Sprintf("%s/32", getNetwork(uint32(i))))

		if err != nil {
			log.Debug("error: ", err)
		}

		_, err = createQDiscNew(tapDevices[tapName], rtnl)

		if err != nil {
			log.Debugf("error: %v", err)
		}

		pbar.Add(1)

	}

	log.Debug("finished cleaning and creating tap devices")

	return devId, tapDevices
}

func timeNewSync(N int, f *os.File) {
	cfg := tc.Config{
		NetNS:  0,
		Logger: nil,
	}

	rtnl, err := tc.Open(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open rtnetlink socket: %v\n", err)
		return
	}
	defer func() {
		if err := rtnl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close rtnetlink socket: %v\n", err)
		}
	}()

	_, tapDevices := cleanNew(N, rtnl)

	// pbar := bar.New((N - 2) * (N - 2) / 2)
	pbar := bar.New(N * (N - 1))

	start := time.Now()

	// both for-loops now cover an upper triangle (every possible unordered pairing of IPs, excl self)
	for i := 0; i < N; i++ {
		//tapName := "tap" + strconv.Itoa(i)

		for j := N - 1; j >= 0; j-- {
			// skip self
			if j == i {
				continue
			}

			a := getNetwork(uint32(i))
			b := getNetwork(uint32(j))

			innerStart := time.Now()

			//index := uint32(i*N + j)
			index := uint32(j + 2)

			err = createLinkNew(tapDevices[fmt.Sprintf("tap%d", i)], rtnl, index, a, b, 1.1)

			if err != nil {
				log.Debugf("error: ", err)
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

	log.Debugf("finished new process in %s", elapsed)
}

func timeNewAsync(N int, f *os.File) {
	var wg sync.WaitGroup

	cfg := tc.Config{
		NetNS:  0,
		Logger: nil,
	}

	rtnl, err := tc.Open(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open rtnetlink socket: %v\n", err)
		return
	}

	_, tapDevices := cleanNew(N, rtnl)

	if err := rtnl.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "could not close rtnetlink socket: %v\n", err)
	}

	// results writer
	res := make(chan string)
	die := make(chan struct{})
	go func(f *os.File, res <-chan string, die <-chan struct{}) {
		for {
			select {
			case result := <-res:
				f.WriteString(result)
			case <-die:
				return
			}
		}
	}(f, res, die)

	// worker pooling
	pbar := bar.New(N * (N - 1))

	start := time.Now()

	for i := 0; i < N; i++ {
		wg.Add(1)
		go innerLoopAsync(&wg, tapDevices[fmt.Sprintf("tap%d", i)], pbar, i, N, res)
	}
	wg.Wait()
	die <- struct{}{}

	elapsed := time.Since(start)

	log.Debugf("finished new async process in %s", elapsed)
}

func innerLoopAsync(wg *sync.WaitGroup, devId int, pbar *bar.ProgressBar, i int, N int, res chan<- string) {
	defer wg.Done()
	cfg := tc.Config{
		NetNS:  0,
		Logger: nil,
	}

	rtnl, err := tc.Open(&cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open rtnetlink socket: %v\n", err)
		return
	}
	defer func() {
		if err := rtnl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close rtnetlink socket: %v\n", err)
		}
	}()

	for j := 0; j < N; j++ {
		// skip self

		a := getNetwork(uint32(i))
		b := getNetwork(uint32(j))

		if j == i {
			continue
		}

		// index := uint32(k*N + j + 1)
		innerStart := time.Now()
		index := uint32(j + 2)

		err = createLinkNew(devId, rtnl, index, a, b, 1.1)

		if err != nil {
			log.Debugf("innerLoopAsync: could not create link. error: %v", err)
		}

		/*err = updateDelayNew(devId, rtnl, index, 1.1, 5)

		if err != nil {
			log.Debugf("innerLoopAsync: could not update delay. error: %v", err)
		}*/
		innerEnd := time.Now()
		res <- fmt.Sprintf("%d,%d,%d,%d,%d\n", N, index, i, j, innerEnd.Sub(innerStart).Nanoseconds())

		pbar.Add(1)

		if err != nil {
			log.Debugf("error: ", err)
		}
	}

}

func createBridge() (int, error) {
	log.Debugf("creating br0")

	cmd := exec.Command(CRBR)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return -1, errors.Wrapf(err, "%#v: output: %s", cmd.Args, out)
	}

	// generate device id for network interface from script output
	lineSplit := strings.Split(string(out), "\n")
	outSplit := strings.Split(lineSplit[len(lineSplit)-2], ": ")
	devId, err := strconv.Atoi(outSplit[0])

	log.Debugf("successfully created br0, device ID %d", devId)

	return devId, nil
}

func createTap(tapName string, addr string) (int, error) {
	log.Debugf("creating tap %s", tapName)

	cmd := exec.Command(CRTAP, tapName, BRIDGEINTERFACE, addr)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return 0, errors.Wrapf(err, "%#v: %s", cmd.Args, string(out))
	}

	// generate device id for network interface from script output
	lineSplit := strings.Split(string(out), "\n")
	outSplit := strings.Split(lineSplit[len(lineSplit)-2], ": ")
	devId, err := strconv.Atoi(outSplit[0])

	log.Debugf("successfully created tap %s, device ID %d", tapName, devId)

	return devId, nil
}

// returns a network in CIDR notation for a tap device with index id
func getNetwork(id uint32) string {
	id = id + 1
	// use a /32 subnet for the tap device
	return fmt.Sprintf("10.%d.%d.%d", (id >> 14 & 0xFF), (id >> 6 & 0xFF), (id << 2 & 0xFF))
}
