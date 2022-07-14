package main

import (
	"strconv"
	"strings"

	"github.com/florianl/go-tc"
	helper "github.com/florianl/go-tc/core"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func parseIp(ip string) uint32 {
	// split ip address
	split := strings.Split(ip, ".")
	ipParts := make([]uint32, 4)

	// parse into uint32
	for i, s := range split {
		num, err := strconv.Atoi(s)

		if err != nil {
			log.Debugf("error: ", err)
		}

		ipParts[i] = uint32(num)
	}

	// build into one number
	return ((ipParts[3] << 24) & 0xff000000) | ((ipParts[2] << 16) & 0x00ff0000) | ((ipParts[1] << 8) & 0x0000ff00) | (ipParts[0] & 0x000000ff)
}

func createQDiscNew(devId int, rtnl *tc.Tc) (int64, error) {
	// remove old stuff first, didn't work over rtnl
	err := removeRootQDisc(true)

	if err != nil {
		return 0, errors.WithStack(err)
	}

	// tc qdisc add dev [TAP_NAME] root handle 1: htb default 1 r2q 1
	//cmd := exec.Command(TC, "qdisc", "add", "dev", tapName, "root", "handle", "1:", "htb", "default", "1", "r2q", "1")
	qdisc := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devId),
			Handle:  helper.BuildHandle(0x1, 0x0),
			Parent:  tc.HandleRoot,
			Info:    0,
		},
		Attribute: tc.Attribute{
			Kind: "htb",
			Htb: &tc.Htb{
				Init: &tc.HtbGlob{
					Version:      0x3,
					Rate2Quantum: 0x1,
					Defcls:       0x1,
				},
			},
		},
	}
	err = rtnl.Qdisc().Add(&qdisc)

	if err != nil {
		log.Debugf("createQDiscNew: qdisc adding failed. error: %v", err)
	}

	// tc class add dev [TAP_NAME] parent 1: classid 1:1 htb rate [DEFAULTRATE] quantum 1514
	//cmd = exec.Command(TC, "class", "add", "dev", tapName, "parent", "1:", "classid", "1:1", "htb", "rate", DEFAULTRATE, "quantum", "1514")
	class := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devId),
			Handle:  helper.BuildHandle(0x1, 0x1),
			Parent:  helper.BuildHandle(0x1, 0x0),
		},
		Attribute: tc.Attribute{
			Kind: "htb",
			Htb: &tc.Htb{
				Init: &tc.HtbGlob{
					Version: 0x3,
				},
				Parms: &tc.HtbOpt{
					Rate: tc.RateSpec{
						Linklayer: 0x1,
						Rate:      0xffffffff,
					},
					Ceil: tc.RateSpec{
						Linklayer: 0x1,
						Rate:      0xffffffff,
					},
					Quantum: 0x5ea, // 1514 in hex
				},
			},
		},
	}
	err = rtnl.Class().Add(&class)

	if err != nil {
		log.Debugf("createQDiscNew: class adding failed. error: %v", err)
	}

	// return starting index for delays
	return 1, nil
}

func createLinkNew(devId int, rtnl *tc.Tc, index uint32, a string, b string, delay float64) error {
	// tc class add dev [TAP_NAME] parent 1: classid 1:[INDEX] htb rate [DEFAULTRATE] quantum 1514

	//cmd := exec.Command(TC, "class", "add", "dev", tapName, "parent", "1:", "classid", fmt.Sprintf("1:%x", index), "htb", "rate", DEFAULTRATE, "quantum", "1514")
	class := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devId),
			Handle:  helper.BuildHandle(0x1, index),
			Parent:  helper.BuildHandle(0x1, 0x0),
		},
		Attribute: tc.Attribute{
			Kind: "htb",
			Htb: &tc.Htb{
				Init: &tc.HtbGlob{
					Version: 0x3,
				},
				Parms: &tc.HtbOpt{
					Rate: tc.RateSpec{
						Linklayer: 0x1,
						Rate:      0xffffffff,
					},
					Ceil: tc.RateSpec{
						Linklayer: 0x1,
						Rate:      0xffffffff,
					},
					Quantum: 0x5ea, // 1514 in hex
				},
			},
		},
	}
	err := rtnl.Class().Add(&class)

	if err != nil {
		log.Debugf("createLinkNew: could not add class %d. error: %v", index, err)
	}

	// tc qdisc add dev [TAP_NAME] parent 1:[INDEX] handle [INDEX]: netem delay 0.0ms limit 1000000

	//cmd := exec.Command(TC, "qdisc", "add", "dev", tapName, "parent", fmt.Sprintf("1:%x", index), "handle", fmt.Sprintf("%x:", index), "netem", "delay", "0.0", "limit", "1000000")
	delayNanosec := int64(delay * 1000000)

	qdisc := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devId),
			Handle:  helper.BuildHandle(index, 0x0),
			Parent:  helper.BuildHandle(0x1, index),
			Info:    0,
		},
		Attribute: tc.Attribute{
			Kind: "netem",
			Netem: &tc.Netem{
				Qopt: tc.NetemQopt{
					Limit: 0xf4240, // limit 1_000_000
				},
				Latency64: &delayNanosec,
			},
		},
	}
	err = rtnl.Qdisc().Add(&qdisc)

	if err != nil {
		log.Debugf("createLinkNew: could not add qdisc %d. error: %v", index, err)
	}

	// yes its required to set "src" as "dest_net" and "dst" as "source_net", this is intentional
	// removing the "match ip dst [SOURCE_NET]" filter as it wouldn't do anything: there is only one network on this tap anyway
	// tc filter add dev [TAP_NAME] protocol ip parent 1: prio [INDEX] u32 match ip src [DEST_NET] classid 1:[INDEX]

	//cmd := exec.Command(TC, "filter", "add", "dev", tapName, "protocol", "ip", "parent", "1:", "prio", fmt.Sprintf("%d", index), "u32", "match", "ip", "src", b, "classid", fmt.Sprintf("1:%x", index))

	// TODO: COMMENTED OUT FOR TESTING PURPOSES!!!!!!!!!! later replaced by one ebpf classifier

	// classId := helper.BuildHandle(0x1, 0x0)
	// //chain := uint32(0)
	// //flags := uint32(8)
	// //pcnt := uint64(0)
	// filter := tc.Object{
	// 	Msg: tc.Msg{
	// 		Family:  unix.AF_UNSPEC,
	// 		Ifindex: uint32(devId),
	// 		Handle:  helper.BuildHandle(index, 0x0),
	// 		Parent:  helper.BuildHandle(0x1, index),
	// 		Info:    768,
	// 	},
	// 	Attribute: tc.Attribute{
	// 		Kind: "u32",
	// 		//Chain: &chain,
	// 		U32: &tc.U32{
	// 			ClassID: &classId,
	// 			Sel: &tc.U32Sel{
	// 				Flags: 0x1,
	// 				NKeys: 0x1,
	// 				Keys: []tc.U32Key{
	// 					{
	// 						Mask: 0xffffffff,
	// 						Val:  parseIp(b), // this is the IP address
	// 						Off:  0xc,
	// 					},
	// 				},
	// 			},
	// 			//Flags: &flags,
	// 			//Pcnt:  &pcnt,
	// 		},
	// 	},
	// }
	// err = rtnl.Filter().Add(&filter)

	// if err != nil {
	// 	log.Debugf("createLinkNew: could not add filter %d. error: %v", index, err)
	// }

	return nil
}

func updateDelayNew(devId int, rtnl *tc.Tc, index uint32, delay float64, bandwidth int) error {
	// tc qdisc change dev [TAP_NAME] parent 1:[INDEX] handle [INDEX]: netem delay [DELAY].0ms limit 1000000
	delayNanosec := int64(delay * 1000000)

	//cmd := exec.Command(TC, "qdisc", "change", "dev", tapName, "parent", fmt.Sprintf("1:%x", index), "handle", fmt.Sprintf("%x:", index), "netem", "delay", fmt.Sprintf("%.1fms", delay), "limit", "1000000")
	qdisc := tc.Object{
		Msg: tc.Msg{
			Family:  unix.AF_UNSPEC,
			Ifindex: uint32(devId),
			Handle:  helper.BuildHandle(index, 0x0),
			Parent:  helper.BuildHandle(0x1, index),
			Info:    0,
		},
		Attribute: tc.Attribute{
			Kind: "netem",
			Netem: &tc.Netem{
				Qopt: tc.NetemQopt{
					Limit: 0xf4240, // limit 1_000_000
				},
				Latency64: &delayNanosec,
			},
		},
	}
	err := rtnl.Qdisc().Change(&qdisc)

	if err != nil {
		log.Debugf("updateDelayNew: could not change qdisc %d. error: %v", index, err)
	}

	return nil
}
