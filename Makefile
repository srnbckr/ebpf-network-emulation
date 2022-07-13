all: map-populator da-example ebpf-network-simulation ebpf-delay experiment

da-example:
	go generate ./cmd/da-example
	go build -o bin/da-example ./cmd/da-example

map-populator:
	go build -o bin/map-populator ./cmd/map-populator

ebpf-network-simulation:
	go generate ./cmd/ebpf-network-simulation
	go build -o bin/ebpf-network-simulation ./cmd/ebpf-network-simulation

ebpf-delay:
	go generate ./cmd/ebpf-delay
	go build -o bin/ebpf-delay ./cmd/ebpf-delay

experiment:
	go build -o bin/experiment ./cmd/experiment