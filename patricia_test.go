package main

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"

	"github.com/kentik/patricia"
	"github.com/kentik/patricia/string_tree"
)

// BenchmarkPatriciaInsert benchmarks insertion performance
func BenchmarkPatriciaInsert(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []string
		isV6     bool
	}{
		{
			name: "single_prefix_v4",
			prefixes: []string{
				"192.168.1.0/24",
			},
			isV6: false,
		},
		{
			name: "10_prefixes_v4",
			prefixes: []string{
				"10.0.0.0/8", "10.1.0.0/16", "10.1.1.0/24",
				"192.168.0.0/16", "192.168.1.0/24",
				"172.16.0.0/12", "172.16.1.0/24",
				"8.8.8.0/24", "1.1.1.0/24", "4.4.4.0/24",
			},
			isV6: false,
		},
		{
			name: "100_prefixes_v4",
			prefixes: func() []string {
				var prefixes []string
				for i := range 100 {
					prefixes = append(prefixes,
						fmt.Sprintf("10.%d.0.0/16", i%256))
				}
				return prefixes
			}(),
			isV6: false,
		},
		{
			name: "overlapping_prefixes_v4",
			prefixes: []string{
				"10.0.0.0/8",
				"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16",
				"10.1.1.0/24", "10.1.2.0/24", "10.1.3.0/24",
				"10.1.1.1/32", "10.1.1.2/32", "10.1.1.3/32",
			},
			isV6: false,
		},
		{
			name: "ipv6_prefixes_v6",
			prefixes: []string{
				"2001:db8::/32",
				"2001:db8:1::/48",
				"2001:db8:2::/48",
				"2001:db8:1:1::/64",
			},
			isV6: true,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if bm.isV6 {
					tree := string_tree.NewTreeV6()
					for j, cidr := range bm.prefixes {
						p := netip.MustParsePrefix(cidr)
						addr := p.Addr()
						bits := p.Bits()
						_, _ = tree.Set(patricia.NewIPv6Address(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
					}
				} else {
					tree := string_tree.NewTreeV4()
					for j, cidr := range bm.prefixes {
						p := netip.MustParsePrefix(cidr)
						addr := p.Addr()
						bits := p.Bits()
						_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
					}
				}
			}
		})
	}
}

// BenchmarkPatriciaLookup benchmarks lookup performance
func BenchmarkPatriciaLookup(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []string
		lookups  []string
		isV6     bool
	}{
		{
			name: "single_prefix_match_v4",
			prefixes: []string{
				"192.168.1.0/24",
			},
			lookups: []string{
				"192.168.1.1",
			},
			isV6: false,
		},
		{
			name: "10_prefixes_various_matches_v4",
			prefixes: []string{
				"10.0.0.0/8", "10.1.0.0/16", "10.1.1.0/24",
				"192.168.0.0/16", "192.168.1.0/24",
				"172.16.0.0/12", "8.8.8.0/24",
			},
			lookups: []string{
				"10.0.0.1", "10.1.0.1", "10.1.1.1",
				"192.168.1.1", "172.16.1.1", "8.8.8.8",
			},
			isV6: false,
		},
		{
			name: "100_prefixes_deep_lookup_v4",
			prefixes: func() []string {
				var prefixes []string
				for i := range 100 {
					prefixes = append(prefixes,
						fmt.Sprintf("10.%d.0.0/16", i%256))
				}
				return prefixes
			}(),
			lookups: []string{
				"10.50.0.1", "10.99.0.1", "10.0.0.1",
			},
			isV6: false,
		},
		{
			name: "no_match_v4",
			prefixes: []string{
				"192.168.1.0/24",
			},
			lookups: []string{
				"8.8.8.8",
			},
			isV6: false,
		},
		{
			name: "longest_prefix_match_v4",
			prefixes: []string{
				"10.0.0.0/8",
				"10.1.0.0/16",
				"10.1.1.0/24",
				"10.1.1.128/25",
			},
			lookups: []string{
				"10.1.1.129",
			},
			isV6: false,
		},
		{
			name: "ipv6_lookup_v6",
			prefixes: []string{
				"2001:db8::/32",
				"2001:db8:1::/48",
			},
			lookups: []string{
				"2001:db8:1::1",
			},
			isV6: true,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			if bm.isV6 {
				tree := string_tree.NewTreeV6()
				for j, cidr := range bm.prefixes {
					p := netip.MustParsePrefix(cidr)
					addr := p.Addr()
					bits := p.Bits()
					_, _ = tree.Set(patricia.NewIPv6Address(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
				}

				addrs := make([]netip.Addr, len(bm.lookups))
				for i, s := range bm.lookups {
					addrs[i] = netip.MustParseAddr(s)
				}

				b.ResetTimer()
				b.ReportAllocs()

				for b.Loop() {
					for _, addr := range addrs {
						_, _ = tree.FindDeepestTag(patricia.NewIPv6Address(addr.AsSlice(), 128))
					}
				}
			} else {
				tree := string_tree.NewTreeV4()
				for j, cidr := range bm.prefixes {
					p := netip.MustParsePrefix(cidr)
					addr := p.Addr()
					bits := p.Bits()
					_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
				}

				addrs := make([]netip.Addr, len(bm.lookups))
				for i, s := range bm.lookups {
					addrs[i] = netip.MustParseAddr(s)
				}

				b.ResetTimer()
				b.ReportAllocs()

				for b.Loop() {
					for _, addr := range addrs {
						_, _ = tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
					}
				}
			}
		})
	}
}

// BenchmarkPatriciaInsertAndLookup benchmarks combined insert and lookup
func BenchmarkPatriciaInsertAndLookup(b *testing.B) {
	prefixes := make([]string, 1000)
	for i := range 1000 {
		prefixes[i] = fmt.Sprintf("10.%d.%d.0/24", i/256, i%256)
	}

	b.ReportAllocs()

	for b.Loop() {
		tree := string_tree.NewTreeV4()

		// Insert
		for j, cidr := range prefixes {
			p := netip.MustParsePrefix(cidr)
			addr := p.Addr()
			bits := p.Bits()
			_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
		}

		// Lookup
		for j := range 100 {
			addr := netip.MustParseAddr(fmt.Sprintf("10.%d.%d.1", j/256, j%256))
			_, _ = tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
		}
	}
}

// BenchmarkPatriciaMemoryFootprint measures memory usage
func BenchmarkPatriciaMemoryFootprint(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("prefixes_%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				tree := string_tree.NewTreeV4()

				for j := range size {
					p := netip.MustParsePrefix(
						fmt.Sprintf("10.%d.%d.0/24", j/256, j%256))
					addr := p.Addr()
					bits := p.Bits()
					_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", j))
				}

				// Force allocation tracking
				_ = tree
			}
		})
	}
}

// BenchmarkPatriciaConcurrentLookup benchmarks concurrent lookups
func BenchmarkPatriciaConcurrentLookup(b *testing.B) {
	tree := string_tree.NewTreeV4()

	// Setup with 100 prefixes
	for i := range 100 {
		p := netip.MustParsePrefix(fmt.Sprintf("10.%d.0.0/16", i%256))
		addr := p.Addr()
		bits := p.Bits()
		_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), fmt.Sprintf("DC%d", i))
	}

	addrs := make([]netip.Addr, 100)
	for i := range 100 {
		addrs[i] = netip.MustParseAddr(fmt.Sprintf("10.%d.0.1", i%256))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		rng := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			addr := addrs[rng.Intn(len(addrs))]
			_, _ = tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
		}
	})
}
