struct handle_bps_delay {
    __u32 tc_handle;
    __u32 throttle_rate_bps;
    __u32 delay_ms;
} HANDLE_BPS_DELAY;

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, HANDLE_BPS_DELAY);
    __uint(pinning, LIBBPF_PIN_BY_NAME); // pin map by name (accessible under /sys/fs/bpf/<name>)
    __uint(max_entries, 65535);
} IP_HANDLE_BPS_DELAY SEC(".maps");