#include <stdint.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/stddef.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "maps.h"
#include "helpers.h"

#define NS_PER_MS 1000000

static inline int set_delay(struct __sk_buff *skb, uint32_t *delay_ms) {
    uint64_t delay_ns;
    uint64_t now = bpf_ktime_get_ns();
    delay_ns = (*delay_ms) * NS_PER_MS;
    skb->tstamp = now + delay_ns;
    return TC_ACT_OK;
}

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
    if (eth_type == bpf_htons(ETH_P_IP)) {
        ip_type = parse_iphdr(&nh, data_end, &iphdr);
        if (ip_type == IPPROTO_ICMP || ip_type == IPPROTO_TCP || ip_type == IPPROTO_UDP) {
            __u32 ip_address = iphdr->daddr; // destination IP, to be used as map lookup key
            __u32 *throttle_rate_bps;
            __u32 *delay_ms;
            struct handle_bps_delay *val_struct;
            // Map lookup
            val_struct = bpf_map_lookup_elem(&IP_HANDLE_BPS_DELAY, &ip_address);

            // Safety check, go on if no handle could be retrieved
            if (!val_struct) {
                return TC_ACT_OK;
            }

            delay_ms = &val_struct->delay_ms;
            // Safety check, go on if no handle could be retrieved
            if (!delay_ms) {
                return TC_ACT_OK;
            }

            return set_delay(skb, delay_ms);
        }
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";