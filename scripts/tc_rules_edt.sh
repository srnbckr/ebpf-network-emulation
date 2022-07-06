#!/bin/bash

# cleanup
#tc qdisc del dev wlp2s0 root

# create qdiscs and classes
#tc qdisc add dev wlp2s0 root handle 1a1e: htb default 1
#tc qdisc add dev wlp2s0 clsact

#tc class add dev wlp2s0 parent 1a1e: classid 1a1e:1 htb rate 32000000.0kbit
#tc class add dev wlp2s0 parent 1a1e: classid 1a1e:3 htb rate 32000000.0Kbit ceil 32000000.0Kbit
#tc qdisc add dev wlp2s0 parent 1a1e:3 handle 205d: netem delay 200.0ms

# load the eBPF filter/classifier
#tc filter add dev wlp2s0 parent 1a1e: bpf obj tc_test.o sec cls
#tc filter add dev wlp2s0 ingress bpf da obj tc_test_da.o sec cls
#tc filter add dev wlp2s0 egress bpf da obj tc_test_da.o sec cls

# load u32 filter matching destination IP 1.1.1.1 (for debugging, should be commented out)
#tc filter add dev wlan0 protocol ip parent 1a1e: prio 5 u32 match ip dst 1.1.1.1/32 flowid 1a1e:3


#######
tc qdisc del dev wlp2s0 root
tc qdisc del dev wlp2s0 clsact
tc qdisc add dev wlp2s0 root fq ce_threshold 4ms
tc qdisc add dev wlp2s0 clsact
tc filter add dev wlp2s0 egress bpf obj tc_edt_bandwidth.o sec cls direct-action

