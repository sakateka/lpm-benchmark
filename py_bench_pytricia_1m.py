#!/usr/bin/env python3
import argparse
import gc
import os
import random
import sys
import time
from ipaddress import IPv4Address, IPv6Address

import psutil
import pytricia


def _build_ipv4_prefix(i: int, rng: random.Random) -> tuple[str, str]:
    a = (i >> 16) & 0xFF
    b = (i >> 8) & 0xFF
    c = i & 0xFF
    d = rng.randint(0, 255)
    prefix_len = 8 + rng.randint(0, 24)  # 8..32 inclusive
    return f"{a}.{b}.{c}.{d}/{prefix_len}", f"DC{i}"


def _build_ipv6_prefix(i: int, rng: random.Random) -> tuple[str, str]:
    # Base 2001:db8::/32 with i spread across bytes, remaining random
    b4 = [
        0x20,
        0x01,
        0x0D,
        0xB8,
        (i >> 24) & 0xFF,
        (i >> 16) & 0xFF,
        (i >> 8) & 0xFF,
        i & 0xFF,
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
        rng.randint(0, 255),
    ]
    addr = IPv6Address(bytes(b4))
    prefix_len = 32 + rng.randint(0, 96)  # 32..128 inclusive
    return f"{addr.compressed}/{prefix_len}", f"DC{i}"


def _gen_ipv4_lookup_addrs(num: int, seed: int) -> list[str]:
    rng = random.Random(seed)
    addrs = []
    for _ in range(num):
        addrs.append(
            str(
                IPv4Address(
                    bytes(
                        [
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                        ]
                    )
                )
            )
        )
    return addrs


def _gen_ipv6_lookup_addrs(num: int, seed: int) -> list[str]:
    rng = random.Random(seed)
    addrs = []
    for _ in range(num):
        addrs.append(
            str(
                IPv6Address(
                    bytes(
                        [
                            0x20,
                            0x01,
                            0x0D,
                            0xB8,
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                            rng.randint(0, 255),
                        ]
                    )
                )
            )
        )
    return addrs


def benchmark_pytricia(
    family: str,
    num_prefixes: int,
    lookup_probes: int,
    lookup_set_size: int,
) -> None:
    proc = psutil.Process(os.getpid())
    gc.collect()
    rss_before = proc.memory_info().rss

    if family == "ipv4":
        rng = random.Random(42)
        prefixes = [_build_ipv4_prefix(i, rng) for i in range(num_prefixes)]
        pyt = pytricia.PyTricia(32)
    else:
        rng = random.Random(42)
        prefixes = [_build_ipv6_prefix(i, rng) for i in range(num_prefixes)]
        pyt = pytricia.PyTricia(128)

    # Insert benchmark
    t0 = time.perf_counter()
    for key, val in prefixes:
        pyt[key] = val
    t1 = time.perf_counter()

    gc.collect()
    rss_after = proc.memory_info().rss
    rss_delta = rss_after - rss_before

    insert_elapsed = t1 - t0
    insert_qps = num_prefixes / insert_elapsed if insert_elapsed > 0 else float("inf")

    # Prepare lookup addresses
    if family == "ipv4":
        addrs = _gen_ipv4_lookup_addrs(lookup_set_size, 43)
    else:
        addrs = _gen_ipv6_lookup_addrs(lookup_set_size, 43)

    # Lookup benchmark
    t2 = time.perf_counter()
    found = 0
    for i in range(lookup_probes):
        a = addrs[i % len(addrs)]
        # get_key performs LPM and returns the prefix; None if no match
        k = pyt.get_key(a)
        if k is not None:
            found += 1
    t3 = time.perf_counter()

    lookup_elapsed = t3 - t2
    lookup_qps = lookup_probes / lookup_elapsed if lookup_elapsed > 0 else float("inf")

    insert_ns_per_op = (insert_elapsed * 1e9) / num_prefixes if num_prefixes else float("nan")
    lookup_ns_per_op = (lookup_elapsed * 1e9) / lookup_probes if lookup_probes else float("nan")
    rss_delta_mb = rss_delta / (1024 * 1024)
    rss_after_mb = rss_after / (1024 * 1024)

    print(f"PyTricia {family.upper()} benchmark")
    print(f"  prefixes:         {num_prefixes:,}")
    print(f"  insert:           {insert_elapsed:,.6f} s  |  {insert_qps:,.2f} ops/s  |  {insert_ns_per_op:,.2f} ns/op")
    print(f"  lookup probes:    {lookup_probes:,}")
    print(f"  lookup:           {lookup_elapsed:,.6f} s  |  {lookup_qps:,.2f} ops/s  |  {lookup_ns_per_op:,.2f} ns/op  |  found={found:,}")
    print(f"  RSS after insert: {rss_after:,} B ({rss_after_mb:,.2f} MB)")
    print(f"  RSS delta:        {rss_delta:,} B ({rss_delta_mb:,.2f} MB)")


def main() -> int:
    parser = argparse.ArgumentParser(
        description="PyTricia 1M benchmark: insert, lookup, memory"
    )
    parser.add_argument(
        "--family",
        choices=["ipv4", "ipv6", "both"],
        default="both",
        help="Address family to benchmark",
    )
    parser.add_argument(
        "--count",
        type=int,
        default=1_000_000,
        help="Number of prefixes to insert",
    )
    parser.add_argument(
        "--lookup-probes",
        type=int,
        default=5_000_000,
        help="Number of lookup probes to run",
    )
    parser.add_argument(
        "--lookup-set-size",
        type=int,
        default=1_000,
        help="Distinct random addresses to cycle through for lookups",
    )

    args = parser.parse_args()

    if args.family in ("ipv4", "both"):
        benchmark_pytricia("ipv4", args.count, args.lookup_probes, args.lookup_set_size)
    if args.family in ("ipv6", "both"):
        benchmark_pytricia("ipv6", args.count, args.lookup_probes, args.lookup_set_size)

    return 0


if __name__ == "__main__":
    sys.exit(main())


