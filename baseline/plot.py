#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Thu Jun 16 16:39:51 2022

@author: njapke
"""

import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

df = pd.read_csv("latencies.csv", index_col="Filters")

fig = plt.figure(figsize=(14,4))
ax = plt.axes(xlim=(1, 65534), ylim=(0, 6))
ax.grid(True)
#ax.set_title('latency by filter (matched IP address)')
ax.set_title('latency by filter (no match)')
ax.set_xlabel('number of filters')
ax.set_ylabel('measured latency in ms')

ax.plot(np.array(df.index), np.array(df["Avg Latency"]))

fig.savefig("./latency.png")

df_bnd = pd.read_csv("bandwidths.csv", index_col="Filters")

fig = plt.figure(figsize=(14,4))
ax = plt.axes(xlim=(1, 65534), ylim=(0, 8))
ax.grid(True)
#ax.set_title('bandwidth by filter (matched IP address)')
ax.set_title('bandwidth by filter (no match)')
ax.set_xlabel('number of filters')
ax.set_ylabel('measured bandwidth in Gbit/s')

ax.plot(np.array(df_bnd.index), np.array(df_bnd["Bitrate"]))

fig.savefig("./bandwidth.png")

