#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Jun 16 15:18:09 2022

@author: njapke
"""

import pandas as pd
import os

latencies = []

# parse latencies
for fname in os.listdir("./ping_res"):
    num_filters = int(fname)
    with open(os.path.join("./ping_res", fname), "r") as f:
        flines = f.readlines()
        split_on_space = flines[-1].split()
        split_on_slash = split_on_space[3].split("/")
        min_lat = float(split_on_slash[0])
        avg_lat = float(split_on_slash[1])
        max_lat = float(split_on_slash[2])
        latencies.append(
            {
                "Filters": num_filters,
                "Min Latency": min_lat,
                "Avg Latency": avg_lat,
                "Max Latency": max_lat
            }
        )

lat_df = pd.DataFrame(latencies)
lat_df.set_index("Filters", inplace=True)
lat_df.sort_index(inplace=True)

bandwidths = []

# parse bandwidths
for fname in os.listdir("./iperf_res"):
    num_filters = int(fname)
    with open(os.path.join("./iperf_res", fname), "r") as f:
        flines = f.readlines()
        split_on_space = flines[-4].split()
        transfer = float(split_on_space[4])
        bitrate = float(split_on_space[6])
        if split_on_space[7] == 'Mbits/sec':
            transfer = transfer/1000
            bitrate = bitrate/1000
        bandwidths.append(
            {
                "Filters": num_filters,
                "Transfer": transfer,
                "Bitrate": bitrate
            }
        )

bnd_df = pd.DataFrame(bandwidths)
bnd_df.set_index("Filters", inplace=True)
bnd_df.sort_index(inplace=True)

lat_df.to_csv("./latencies.csv")

bnd_df.to_csv("./bandwidths.csv")

