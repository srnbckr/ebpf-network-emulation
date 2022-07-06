all: map-populator da-example edt-bandwidth-limit ebpf-delay

da-example:
	go generate ./cmd/da-example
	go build -o bin/da-example ./cmd/da-example

map-populator:
	go build -o bin/map-populator ./cmd/map-populator

edt-bandwidth-limit:
	go generate ./cmd/edt-bandwidth-limit
	go build -o bin/edt-bandwidth-limit ./cmd/edt-bandwidth-limit

ebpf-delay:
	go generate ./cmd/ebpf-delay
	go build -o bin/ebpf-delay ./cmd/ebpf-delay