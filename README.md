# ebpf-network-emulation

This repository contains experimental code for a network emulation tool based on eBPF as well as experiment artifacts for it.
We also include experiment setups and results to test NetEm.

## Research

If you use this software in a publication, please cite it as:

### Text

S. Becker, T. Pfandzelter, N. Japke, D. Bermbach, O. Kao, **Network Emulation in Large-Scale Virtual Edge Testbeds: A Note of Caution and the Way Forward**, Proceedings of the 2nd International Workshop on Testing Distributed Internet of Things Systems (TDIS 2022), Pacific Grove, CA, USA, 2022.

### BibTeX

```bibtex
@inproceedings{becker_netem_ebpf:_2022,
    title = "Network Emulation in Large-Scale Virtual Edge Testbeds: A Note of Caution and the Way Forward",
    booktitle = "Proceedings of the 2nd International Workshop on Testing Distributed Internet of Things Systems (TDIS 2022)",
    author = "Becker, Soeren and Pfandzelter, Tobias Pfandzelter and Japke, Nils and Bermbach, David and Kao, Odej",
    year = 2022
}
```

For a full list of publications, please see [our website](https://www.mcc.tu-berlin.de/menue/forschung/publikationen/parameter/en/).

### License

A license for the code in this repository is pending.

## Dependencies

The following installs the needed dependencies on Ubuntu 20.04 / 22.04:

```
sudo apt-get install -y clang gcc-multilib libbpf-dev
```

## Compilation

Use the provided Makefile, the eBPF programs will be compiled with `go generate`:

```sh
make all
```

The binaries can be found in the `bin` dir.

## Examples

### da-example

Simple example showcasing the direct-action flag of tc.

### ebpf-delay

Set different delays per IP destination based on a shared eBPF map and a single `fq` qdisc.

#### Usage

```sh
sudo ./bin/ebpf-delay -iface eth0
```

### ebpf-network-simulation

Limit the available bandwidth to different IP destinations using the [Earliest Departure Time model](https://legacy.netdevconf.info/0x14/pub/slides/55/slides.pdf) and set a delay, based on a shared eBPF map.

#### Usage

```sh
sudo ./bin/ebpf-network-simulation -iface eth0
```

### map-populator

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

#### Usage

First run either `ebpf-delay` or `ebpf-network-simulation`.

```sh
sudo ./bin/map-populator
```

## Remove

```bash
sudo tc qdisc del dev eth0 clsact
sudo tc qdisc del dev eth0 root
sudo rm /sys/fs/bpf/IP_HANDLE_BPS_DELAY
sudo rm /sys/fs/bpf/progs
```

## Used Resources

[njapke/ebpf-ip-classifier](https://github.com/njapke/ebpf-ip-classifier)

[njapke/ebpf-map-populator](https://github.com/njapke/ebpf-map-populator)

[njapke/tc-qdisc-test](https://github.com/njapke/tc-qdisc-test
)

[Replacing HTB with EDT and BPF](https://legacy.netdevconf.info/0x14/session.html?talk-replacing-HTB-with-EDT-and-BPF)
