package main

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"

	"github.com/sakateka/lpm"
)

// BenchmarkLPMInsert benchmarks insertion performance
func BenchmarkLPMInsert(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []string
	}{
		{
			name: "single_prefix",
			prefixes: []string{
				"192.168.1.0/24",
			},
		},
		{
			name: "10_prefixes",
			prefixes: []string{
				"10.0.0.0/8", "10.1.0.0/16", "10.1.1.0/24",
				"192.168.0.0/16", "192.168.1.0/24",
				"172.16.0.0/12", "172.16.1.0/24",
				"8.8.8.0/24", "1.1.1.0/24", "4.4.4.0/24",
			},
		},
		{
			name: "100_prefixes",
			prefixes: func() []string {
				var prefixes []string
				for i := range 100 {
					prefixes = append(prefixes,
						fmt.Sprintf("10.%d.0.0/16", i%256))
				}
				return prefixes
			}(),
		},
		{
			name: "overlapping_prefixes",
			prefixes: []string{
				"10.0.0.0/8",
				"10.1.0.0/16", "10.2.0.0/16", "10.3.0.0/16",
				"10.1.1.0/24", "10.1.2.0/24", "10.1.3.0/24",
				"10.1.1.1/32", "10.1.1.2/32", "10.1.1.3/32",
			},
		},
		{
			name: "ipv6_prefixes",
			prefixes: []string{
				"2001:db8::/32",
				"2001:db8:1::/48",
				"2001:db8:2::/48",
				"2001:db8:1:1::/64",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				lpm := lpm.New()
				for j, cidr := range bm.prefixes {
					prefix := netip.MustParsePrefix(cidr)
					lpm.Insert(prefix, fmt.Sprintf("DC%d", j))
				}
			}
		})
	}
}

// BenchmarkLPMLookup benchmarks lookup performance
func BenchmarkLPMLookup(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []string
		lookups  []string
	}{
		{
			name: "single_prefix_match",
			prefixes: []string{
				"192.168.1.0/24",
			},
			lookups: []string{
				"192.168.1.1",
			},
		},
		{
			name: "10_prefixes_various_matches",
			prefixes: []string{
				"10.0.0.0/8", "10.1.0.0/16", "10.1.1.0/24",
				"192.168.0.0/16", "192.168.1.0/24",
				"172.16.0.0/12", "8.8.8.0/24",
			},
			lookups: []string{
				"10.0.0.1", "10.1.0.1", "10.1.1.1",
				"192.168.1.1", "172.16.1.1", "8.8.8.8",
			},
		},
		{
			name: "100_prefixes_deep_lookup",
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
		},
		{
			name: "no_match",
			prefixes: []string{
				"192.168.1.0/24",
			},
			lookups: []string{
				"8.8.8.8",
			},
		},
		{
			name: "longest_prefix_match",
			prefixes: []string{
				"10.0.0.0/8",
				"10.1.0.0/16",
				"10.1.1.0/24",
				"10.1.1.128/25",
			},
			lookups: []string{
				"10.1.1.129", // Should match /25
			},
		},
		{
			name: "ipv6_lookup",
			prefixes: []string{
				"2001:db8::/32",
				"2001:db8:1::/48",
			},
			lookups: []string{
				"2001:db8:1::1",
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Setup
			lpm := lpm.New()
			for j, cidr := range bm.prefixes {
				prefix := netip.MustParsePrefix(cidr)
				lpm.Insert(prefix, fmt.Sprintf("DC%d", j))
			}

			addrs := make([]netip.Addr, len(bm.lookups))
			for i, lookup := range bm.lookups {
				addrs[i] = netip.MustParseAddr(lookup)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				for _, addr := range addrs {
					_, _ = lpm.Lookup(addr)
				}
			}
		})
	}
}

// BenchmarkLPMInsertAndLookup benchmarks combined insert and lookup
func BenchmarkLPMInsertAndLookup(b *testing.B) {
	prefixes := make([]string, 1000)
	for i := range 1000 {
		prefixes[i] = fmt.Sprintf("10.%d.%d.0/24", i/256, i%256)
	}

	b.ReportAllocs()

	for b.Loop() {
		lpm := lpm.New()

		// Insert
		for j, cidr := range prefixes {
			prefix := netip.MustParsePrefix(cidr)
			lpm.Insert(prefix, fmt.Sprintf("DC%d", j))
		}

		// Lookup
		for j := range 100 {
			addr := netip.MustParseAddr(fmt.Sprintf("10.%d.%d.1", j/256, j%256))
			_, _ = lpm.Lookup(addr)
		}
	}
}

// BenchmarkLPMMemoryFootprint measures memory usage
func BenchmarkLPMMemoryFootprint(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("prefixes_%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				lpm := lpm.New()

				for j := range size {
					prefix := netip.MustParsePrefix(
						fmt.Sprintf("10.%d.%d.0/24", j/256, j%256))
					lpm.Insert(prefix, fmt.Sprintf("DC%d", j))
				}

				// Force allocation tracking
				_ = lpm.Stats()
			}
		})
	}
}

// BenchmarkLPMConcurrentLookup benchmarks concurrent lookups
func BenchmarkLPMConcurrentLookup(b *testing.B) {
	lpm := lpm.New()

	// Setup with 100 prefixes
	for i := range 100 {
		prefix := netip.MustParsePrefix(fmt.Sprintf("10.%d.0.0/16", i%256))
		lpm.Insert(prefix, fmt.Sprintf("DC%d", i))
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
			_, _ = lpm.Lookup(addr)
		}
	})
}
