

// hdr_cursor is used to keep track of the current position in data parsing
struct hdr_cursor {
    void *pos;
};

// parse_ethhdr parses the ethernet header of a packet, and performs necessary bounds checks.
// returns the next protocol
static __always_inline int parse_ethhdr(struct hdr_cursor *nh,
                                        void *data_end,
                                        struct ethhdr **ethhdr)
{
    struct ethhdr *eth = nh->pos;
    int hdrsize = sizeof(*eth);

    /* Byte-count bounds check; check if current pointer + size of header
     * is after data_end.
     */
    if (nh->pos + hdrsize > data_end)
        return TC_ACT_SHOT;

    nh->pos += hdrsize;
    *ethhdr = eth;

    return eth->h_proto; /* network-byte-order */
}

// parse_iphdr parses the IP header of a packet, and performs necessary bounds checks (more complicated due to variable length of IPv4).
// returns the next protocol
static __always_inline int parse_iphdr(struct hdr_cursor *nh,
                                       void *data_end,
                                       struct iphdr **iphdr)
{
    struct iphdr *iph = nh->pos;
    int hdrsize;

    if (iph + 1 > data_end)
        return TC_ACT_SHOT;

    hdrsize = iph->ihl * 4;
    /* Sanity check packet field is valid */
    if(hdrsize < sizeof(*iph))
        return TC_ACT_SHOT;

    /* Variable-length IPv4 header, need to use byte-based arithmetic */
    if (nh->pos + hdrsize > data_end)
        return TC_ACT_SHOT;

    nh->pos += hdrsize;
    *iphdr = iph;

    return iph->protocol;
}