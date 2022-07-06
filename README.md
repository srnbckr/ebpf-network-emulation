# ebpf-network-simulator

# Compilation
Use the provided Makefile, the eBPF programs will be compiled with `go generate`:

```
make all
```

The binaries can be found in the `bin` dir.

# Examples

## da-example
Simple example showcasing the direct-action flag of tc.

## ebpf-delay
Set different delays per IP destination based on a shared eBPF map and a single `fq` qdisc.

### Usage
```
sudo ./bin/ebpf-delay -iface eth0
```

## edt-bandwidth-limit
Limit the available bandwidth to different IP destinations using the [Earliest Departure Time model](https://legacy.netdevconf.info/0x14/pub/slides/55/slides.pdf) and a shared eBPF map.

### Usage
```
sudo ./bin/edt-bandwith-limit -iface eth0
```

## map-populator
Populate an eBPF map with a given number of entries. The map has the following structure and is utilized by `ebpf-delay` and `edt-bandwidth-limit`:


```C
struct handle_bps_delay {
    __u32 tc_handle;          // TC handle, used for classic htb qdisc version
    __u32 throttle_rate_bps;  // Throttled bandwidth (in BPS)
    __u32 delay_ms;           // Delay in ms
} HANDLE_BPS_DELAY;

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);              // The IP destination
    __type(value, HANDLE_BPS_DELAY);
    __uint(pinning, LIBBPF_PIN_BY_NAME); 
    __uint(max_entries, 65535);
} IP_HANDLE_BPS_DELAY SEC(".maps");
```

The map declaration can be found [here](cmd/headers/maps.h).

### Usage
First run either `ebpf-delay` or `edt-bandwidth-limit`.
```
sudo ./bin/map-populator
```