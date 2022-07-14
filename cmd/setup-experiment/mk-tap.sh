#!/bin/bash

NAME=$1
IFACE=$2
NETW=$3

ip link delete "$NAME"

ip tuntap add "$NAME" mode tap

ip addr add "$NETW" dev "$NAME"
ip link set "$NAME" up
sh -c "echo 1 > /proc/sys/net/ipv4/ip_forward"
iptables -t nat -A POSTROUTING -o "$IFACE" -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i "$NAME" -o "$IFACE" -j ACCEPT
ip addr | grep "[0-9]*:\s$NAME:\s"