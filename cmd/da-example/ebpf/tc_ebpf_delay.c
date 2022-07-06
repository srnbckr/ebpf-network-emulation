#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <stdint.h>
#include "maps.h"
#include "helpers.h"

/*
 * This uses tc direct-action mode to set the tc classid with skb->tc_priority
 *  To use on i.e. egress use:
 *      tc qdisc add dev wlp2s0 clsact
 *      tc filter add dev wlp2s0 egress bpf obj tc_test_da.o sec cls direct-action
 *
 * Add htq qdiscs / classes accordingly and adjust the used eBPF map -> see scripts/tc_rules_da.sh
 */
SEC("tc")
int tc_main(struct __sk_buff *skb)
{
    // data_end is a void* to the end of the packet. Needs weird casting due to kernel weirdness.
    void *data_end = (void *)(unsigned long long)skb->data_end;
    // data is a void* to the beginning of the packet. Also needs weird casting.
    void *data = (void *)(unsigned long long)skb->data;

    // nh keeps track of the beginning of the next header to parse
    struct hdr_cursor nh;

    struct ethhdr *eth;
    struct iphdr *iphdr;

    int eth_type;
    int ip_type;

    // start parsing at beginning of data
    nh.pos = data;

    // parse ethernet
    eth_type = parse_ethhdr(&nh, data_end, &eth);

    if (eth_type == bpf_htons(ETH_P_IP)) { // if the next protocol is IPv4
        // parse IPv4
        ip_type = parse_iphdr(&nh, data_end, &iphdr);
        if (ip_type == IPPROTO_ICMP || ip_type == IPPROTO_TCP || ip_type == IPPROTO_UDP) {
            __u32 ip_address = iphdr->daddr; // destination IP, to be used as map lookup key
            __u32 *handle;
            struct handle_bps_delay *val_struct;
            // Map lookup
            val_struct = bpf_map_lookup_elem(&IP_HANDLE_BPS_DELAY, &ip_address);
            //handle = bpf_map_lookup_elem(&IP_TO_HANDLE_MAP, &ip_address);

            // Safety check, go on if no handle could be retrieved
            if (!val_struct) {
                return TC_ACT_OK;
            }
            handle = &val_struct->tc_handle;

            if (!handle) {
                return TC_ACT_OK;
            }

            // set handle as classid
            skb->priority = *handle;
            return TC_ACT_OK;
        }
    }
    // otherwise, use default (flowid given in TC invocation)
    return TC_ACT_OK;
}

// some eBPF kernel features are gated behind the GPL license
char _license[] SEC("license") = "GPL";
