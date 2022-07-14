#!/bin/bash

NAME=$1
BRIDGE=$2
NETW=$3

ip link delete "$NAME"

ip tuntap add "$NAME" mode tap
ip addr add "$NETW" dev "$NAME"
ip link set "$NAME" up
ip link set dev "$NAME" master "$BRIDGE"

ip addr | grep "[0-9]*:\s$NAME:\s"