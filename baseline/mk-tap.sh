#!/bin/bash

NAME=tap1
IP=172.18.0.1

# Delete if already created
ip link delete $NAME

ip tuntap add $NAME mode tap

ip addr add $IP/24 dev $NAME
ip link set $NAME up
sh -c "echo 1 > /proc/sys/net/ipv4/ip_forward"
sh -c "echo 0 > /proc/sys/net/ipv4/conf/$NAME/rp_filter"
iptables -t nat -A POSTROUTING -o wlan0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i $NAME -o wlan0 -j ACCEPT
#ip addr | grep "[0-9]*:\s$NAME:\s"
