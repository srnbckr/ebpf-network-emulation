#!/bin/bash

sh -c "echo 1 > /proc/sys/net/ipv4/ip_forward"

ip link add name br0 type bridge
ip link set dev br0 up

ip address add 192.168.1.1/24 dev br0
iptables -t nat -A POSTROUTING -o br0 -j MASQUERADE
iptables --insert FORWARD --in-interface br0 -j ACCEPT

ip addr | grep "[0-9]*:\sbr0:\s"