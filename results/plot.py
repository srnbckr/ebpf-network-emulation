import seaborn as sns
import pandas as pd
from typing import List
from matplotlib import pyplot as plt
from matplotlib import ticker as tick
from plot_utils import plot_set_font, plot_set_size


PAPER_WIDTH = 516.0


def main():
    # read files
    lat_htb_df = pd.read_csv("latency_htb.csv", index_col="Filters")
    bandwidth_htb_df = pd.read_csv("bandwidth_htb.csv", index_col="Filters")
    lat_ebpf_df = pd.read_csv("latency_ebpf.csv", index_col="Filters")
    bandwidth_ebpf_df = pd.read_csv("bandwidth_ebpf.csv", index_col="Filters")

    # plot single graphs
    plot_latency(
        lat_htb_df,
        title="HTB: latency by filter (no match)",
        filename="latency_htb.pdf",
    )
    plot_bandwidth(
        bandwidth_htb_df,
        title="HTB: bandwidth by filter (no match)",
        filename="bandwidth_htb.pdf",
    )
    plot_latency(
        lat_ebpf_df,
        title="eBPF: latency by filter (no match)",
        filename="latency_ebpf.pdf",
    )
    plot_bandwidth(
        bandwidth_ebpf_df,
        title="eBPF: bandwidth by filter (no match)",
        filename="bandwidth_ebpf.pdf",
    )

    # plot combined graphs
    plot_latency_combined(
        lat_htb_df,
        lat_ebpf_df,
        title="HTB vs eBPF: latency by filter (no match)",
        filename="latency_combined.pdf",
    )
    plot_bandwidth_combined(
        bandwidth_htb_df,
        bandwidth_ebpf_df,
        title="HTB vs eBPF: bandwidth by filter (no match)",
        filename="bandwidth_combined.pdf",
    )


def plot_latency_combined(
    df_htb: pd.DataFrame, df_ebpf: pd.DataFrame, title: str, filename: str
):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat([df_htb, df_ebpf], keys=["HTB", "eBPF"], names=["Method"])
    # print(combined.head())

    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=1))
    g = sns.lineplot(
        data=combined,
        x="Filters",
        y="Avg Latency",
        style="Method",
        hue="Method",
        palette="muted",
        # markers=True,
        ax=axs,
    )
    g.set_title(title)
    g.set_xlabel("Number of filters / map entries")
    g.set_ylabel("Avg latency [ms]")
    g.grid(True)
    g.set_xlim(0, 65534)
    # g.set_ylim(0, 15)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_bandwidth_combined(
    df_htb: pd.DataFrame, df_ebpf: pd.DataFrame, title: str, filename: str
):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat([df_htb, df_ebpf], keys=["HTB", "eBPF"], names=["Method"])
    # print(combined.head())

    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=1))
    g = sns.lineplot(
        data=combined,
        x="Filters",
        y="Bitrate",
        style="Method",
        hue="Method",
        palette="muted",
        # markers=True,
        ax=axs,
    )
    g.set_title(title)
    g.set_xlabel("Number of filters / map entries")
    g.set_ylabel("Measured bandwidth in Gbit/s]")
    g.grid(True)
    g.set_xlim(0, 65534)
    # g.set_ylim(0, 25)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_latency(df: pd.DataFrame, title: str, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    fig, axs = plt.subplots(1, 1, figsize=(14, 4))
    g = sns.lineplot(data=df, x="Filters", y="Avg Latency", palette="muted", ax=axs)
    g.set_title(title)
    g.set_xlabel("number of filters")
    g.set_ylabel("avg latency in ms")
    g.grid(True)
    g.set_xlim(1, 65534)
    # g.set_ylim(0, 20)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_bandwidth(df: pd.DataFrame, title: str, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    fig, axs = plt.subplots(1, 1, figsize=(14, 4))
    g = sns.lineplot(data=df, x="Filters", y="Bitrate", palette="muted", ax=axs)
    g.set_title(title)
    g.set_xlabel("number of filters")
    g.set_ylabel("measured bandwidth in Gbit/s")
    g.grid(True)
    g.set_xlim(1, 65534)
    # g.set_ylim(0, 25)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


if __name__ == "__main__":
    main()
