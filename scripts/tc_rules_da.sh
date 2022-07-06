#!/bin/bash

INTERFACE=wlp2s0

# cleanup
tc qdisc del dev $INTERFACE root
tc qdisc del dev $INTERFACE clsact

# add htb qdiscs
tc qdisc add dev $INTERFACE root handle 1a1e: htb default 1
tc class add dev $INTERFACE parent 1a1e: classid 1a1e:1 htb rate 32000000.0kbit
tc class add dev $INTERFACE parent 1a1e: classid 1a1e:3 htb rate 32000000.0Kbit ceil 32000000.0Kbit
tc qdisc add dev $INTERFACE parent 1a1e:3 handle 205d: netem delay 200.0ms

# add clsact qdisc and attach ebpf prog
tc qdisc add dev $INTERFACE clsact
tc filter add dev $INTERFACE egress bpf obj tc_test_da.o sec cls direct-action


