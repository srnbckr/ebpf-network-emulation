#!/bin/bash

# 1-step compilation does not add debug info (smaller size)
#clang -target bpf -Wall -O2 -c xdp_test.c -o xdp_test.o

# 2-step compilation adds debug info (for some reason), which I need
clang -S \
    -target bpf \
    -Wall \
    -O2 -emit-llvm -c -g -o tc_test.ll tc_test.c
llc -march=bpf -mattr=dwarfris -filetype=obj -o tc_test.o tc_test.ll
