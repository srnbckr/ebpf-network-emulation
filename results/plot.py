import seaborn as sns
import pandas as pd
from typing import List
from matplotlib import pyplot as plt
from matplotlib import ticker as tick
from plot_utils import plot_set_font, plot_set_size


PAPER_WIDTH = 516.0


def main():

    for node in ["gcp", "metal"]:
        # Figure 1
        # Setup experiment
        df_setup = pd.read_csv(f"./{node}/setup-log-1024.csv")
        plot_setup_time_per_link(df_setup, filename=f"time_per_link_rel_{node}.pdf")
        # plot_setup_exp_ecdf(df_setup, filename=f"./graphs/exp_ecdf_rel_{node}.pdf")
        # plot_setup_exp(df_setup, filename=f"./graphs/setup_exp_{node}.pdf")

        # Figure 2 and 3
        # Netem bandwidth and latency for scalability analysis
        latency = pd.read_csv(f"./{node}/latency_netem_nomatch.csv")
        bandwidth = pd.read_csv(f"./{node}/bandwidth_netem_nomatch.csv")
        latency_30000 = pd.read_csv(f"./{node}/latency_netem_match_30000.csv")
        bandwidth_30000 = pd.read_csv(f"./{node}/bandwidth_netem_match_30000.csv")

        plot_latency_htb_matched(
            [latency, latency_30000], filename=f"latency_netem_matches_{node}.pdf"
        )
        plot_bandwidth_htb_matched(
            [bandwidth, bandwidth_30000],
            filename=f"bandwidth_netem_matches_{node}.pdf",
        )

        # Figure 5 and 6
        # netem and eBPF compared
        latency_ebpf = pd.read_csv(f"./{node}/latency_ebpf.csv")
        bandwidth_ebpf = pd.read_csv(f"./{node}/bandwidth_ebpf.csv")

        plot_latency_combined(
            latency,
            latency_ebpf,
            filename=f"latency_combined_{node}.pdf",
        )
        plot_bandwidth_combined(
            bandwidth,
            bandwidth_ebpf,
            filename=f"bandwidth_combined_{node}.pdf",
        )

        # Figure 7
        # ping experiment
        if node == "gcp":
            df = pd.read_csv(f"./{node}/ping_experiment.csv")
            plot_ping_experiment(df, filename=f"ping_experiment_{node}.pdf")


def plot_latency_htb_matched(df_list: List[pd.DataFrame], filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat(df_list, keys=["No Match", "30,000"], names=["Matched Index"])
    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
    g = sns.lineplot(
        data=combined,
        x="Filters",
        y="Avg Latency",
        style="Matched Index",
        hue="Matched Index",
        palette="muted",
        # markers=True,
        ax=axs,
    )
    g.set_xlabel("\#Filters")
    g.set_ylabel("Mean Latency [ms]")
    g.grid(True)
    g.set_xlim(0, 65534)
    # g.set_ylim(0, 15)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_bandwidth_htb_matched(df_list: List[pd.DataFrame], filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat(df_list, keys=["No Match", "30,000"], names=["Matched Index"])
    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
    g = sns.lineplot(
        data=combined,
        x="Filters",
        y="Bitrate",
        style="Matched Index",
        hue="Matched Index",
        palette="muted",
        # markers=True,
        ax=axs,
    )
    g.set_xlabel("\#Filters")
    g.set_ylabel("Throughput [Gbit/s]")
    g.grid(True)
    g.set_xlim(0, 65534)
    # g.set_ylim(0, 15)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_latency_combined(df_htb: pd.DataFrame, df_ebpf: pd.DataFrame, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat([df_htb, df_ebpf], keys=["NetEm", "eBPF"], names=["Method"])
    # print(combined.head())

    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
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
    g.set_xlabel("\#Filters / Map Entries")
    g.set_ylabel("Mean Latency [ms]")
    g.grid(True)
    g.set_xlim(0, 65534)
    # g.set_ylim(0, 15)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_bandwidth_combined(df_htb: pd.DataFrame, df_ebpf: pd.DataFrame, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    combined = pd.concat([df_htb, df_ebpf], keys=["NetEm", "eBPF"], names=["Method"])
    # print(combined.head())

    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
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
    g.set_xlabel("\#Filters / Map Entries")
    g.set_ylabel("Throughput [Gbit/s]")
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


def plot_setup_exp_ecdf(df: pd.DataFrame, filename: str):
    # print(df.head())
    df["time_ms"] = df["time"] / 1e6
    fig, axs = plt.subplots(1, 1, figsize=(14, 4))
    g = sns.ecdfplot(hue="N", x="time_ms", data=df, ax=axs)
    g.set_xlabel("Time per Link (ms)")
    g.set_ylabel("Empirical Cumulative Distribution")
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_setup_exp(df: pd.DataFrame, filename: str):
    df["time_ms"] = df["time"] / 1e6
    fig, axs = plt.subplots(1, 1, figsize=(14, 4))
    g = sns.lineplot(x="index", y="time_ms", hue="N", data=df, ax=axs)
    g.set_ylabel("Time per Link (ms)")
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_setup_time_per_link(df: pd.DataFrame, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
    df["time_ms"] = df["time"] / 1e6
    df["index_2"] = df["i"] * df["N"] + df["j"]
    df["index_rel"] = df["index_2"] / ((df["N"]) * (df["N"] - 1) + (df["N"] - 2))
    df_graph = df
    df_graph["$N$"] = df["N"].astype(str)
    df_graph = df_graph[
        (df_graph["N"] == 128)
        | (df_graph["N"] == 256)
        | (df_graph["N"] == 512)
        | (df_graph["N"] == 1024)
    ]
    df_graph["index_rel_exp"] = df_graph["index_rel"].apply(lambda x: round(x, 2))
    df_graph["index_rel_perc"] = df_graph["index_rel"] * 100
    g = sns.lineplot(
        x="index_rel_perc",
        y="time_ms",
        hue="$N$",
        data=df_graph,
        ci="sd",
        ax=axs,
        palette="muted",
    )
    g.set_ylabel("Time per Link (ms)")
    g.set_xlabel("\% Links")
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


def plot_ping_experiment(df: pd.DataFrame, filename: str):
    sns.set(style="darkgrid")
    plot_set_font()
    # print(df.head())
    data = pd.melt(df, ["sec"])
    data["Method"] = data["variable"]
    # print(data.head())
    # exit(0)
    data = data[data.variable != "baseline"]
    fig, axs = plt.subplots(1, 1, figsize=plot_set_size(PAPER_WIDTH, fraction=0.5))
    g = sns.lineplot(
        data=data,
        x="sec",
        y="value",
        style="Method",
        hue="Method",
        palette="muted",
        # markers=True,
        ax=axs,
    )
    # g.set_title(title)
    g.set_xlabel("Time [s]")
    g.set_ylabel("Latency [ms]")
    g.grid(True)
    # g.set_xlim(0, 65534)
    # g.set_ylim(0, 25)
    plt.tight_layout()
    plt.savefig(filename, format="pdf", bbox_inches="tight")


if __name__ == "__main__":
    main()
