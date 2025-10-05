package main

import (
	"fmt"
	"math/rand"
	"net/netip"
	"runtime"
	"testing"

	"github.com/kentik/patricia"
	"github.com/kentik/patricia/string_tree"
)

// BenchmarkPatriciaInsert1M benchmarks insertion of 1M prefixes
func BenchmarkPatriciaInsert1M(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []netip.Prefix
		values   []string
		isV6     bool
	}{
		{
			name: "ipv4_1M_prefixes",
			prefixes: func() []netip.Prefix {
				rng := rand.New(rand.NewSource(42))
				prefixes := make([]netip.Prefix, 1000_000)
				for i := range prefixes {
					// Generate diverse prefixes with random masks (8-32)
					addr := netip.AddrFrom4([4]byte{
						byte((i >> 16) & 0xff),
						byte((i >> 8) & 0xff),
						byte(i & 0xff),
						byte(rng.Intn(256)),
					})
					prefixLen := 8 + rng.Intn(25)
					prefixes[i] = netip.PrefixFrom(addr, prefixLen).Masked()
				}
				return prefixes
			}(),
			values: func() []string {
				values := make([]string, 1000_000)
				for i := range values {
					values[i] = fmt.Sprintf("DC%d", i)
				}
				return values
			}(),
			isV6: false,
		},
		{
			name: "ipv6_1M_prefixes",
			prefixes: func() []netip.Prefix {
				rng := rand.New(rand.NewSource(42))
				prefixes := make([]netip.Prefix, 1000_000)
				for i := range prefixes {
					// Generate diverse prefixes with random masks (32-128)
					addr := netip.AddrFrom16([16]byte{
						0x20, 0x01, 0x0d, 0xb8,
						byte((i >> 24) & 0xff), byte((i >> 16) & 0xff),
						byte((i >> 8) & 0xff), byte(i & 0xff),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
					})
					prefixLen := 32 + rng.Intn(97)
					prefixes[i] = netip.PrefixFrom(addr, prefixLen).Masked()
				}
				return prefixes
			}(),
			values: func() []string {
				values := make([]string, 1000_000)
				for i := range values {
					values[i] = fmt.Sprintf("DC%d", i)
				}
				return values
			}(),
			isV6: true,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			if bm.isV6 {
				tree := string_tree.NewTreeV6()
				idx := 0

				for b.Loop() {
					addr := bm.prefixes[idx].Addr()
					bits := bm.prefixes[idx].Bits()
					_, _ = tree.Set(patricia.NewIPv6Address(addr.AsSlice(), uint(bits)), bm.values[idx])
					idx = (idx + 1) % 1000_000
				}
			} else {
				tree := string_tree.NewTreeV4()
				idx := 0

				for b.Loop() {
					addr := bm.prefixes[idx].Addr()
					bits := bm.prefixes[idx].Bits()
					_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), bm.values[idx])
					idx = (idx + 1) % 1000_000
				}
			}
		})
	}
}

// BenchmarkPatriciaLookup1M benchmarks lookups in a patricia tree with 1M prefixes
func BenchmarkPatriciaLookup1M(b *testing.B) {
	benchmarks := []struct {
		name     string
		prefixes []netip.Prefix
		values   []string
		addrs    []netip.Addr
		isV6     bool
	}{
		{
			name: "ipv4_1M_prefixes",
			prefixes: func() []netip.Prefix {
				rng := rand.New(rand.NewSource(42))
				prefixes := make([]netip.Prefix, 1000_000)
				for i := range prefixes {
					addr := netip.AddrFrom4([4]byte{
						byte((i >> 16) & 0xff),
						byte((i >> 8) & 0xff),
						byte(i & 0xff),
						byte(rng.Intn(256)),
					})
					prefixLen := 8 + rng.Intn(25)
					prefixes[i] = netip.PrefixFrom(addr, prefixLen).Masked()
				}
				return prefixes
			}(),
			values: func() []string {
				values := make([]string, 1000_000)
				for i := range values {
					values[i] = fmt.Sprintf("DC%d", i)
				}
				return values
			}(),
			addrs: func() []netip.Addr {
				rng := rand.New(rand.NewSource(43))
				addrs := make([]netip.Addr, 1000)
				for i := range addrs {
					addrs[i] = netip.AddrFrom4([4]byte{
						byte(rng.Intn(256)),
						byte(rng.Intn(256)),
						byte(rng.Intn(256)),
						byte(rng.Intn(256)),
					})
				}
				return addrs
			}(),
			isV6: false,
		},
		{
			name: "ipv6_1M_prefixes",
			prefixes: func() []netip.Prefix {
				rng := rand.New(rand.NewSource(42))
				prefixes := make([]netip.Prefix, 1000_000)
				for i := range prefixes {
					addr := netip.AddrFrom16([16]byte{
						0x20, 0x01, 0x0d, 0xb8,
						byte((i >> 24) & 0xff), byte((i >> 16) & 0xff),
						byte((i >> 8) & 0xff), byte(i & 0xff),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
					})
					prefixLen := 32 + rng.Intn(97)
					prefixes[i] = netip.PrefixFrom(addr, prefixLen).Masked()
				}
				return prefixes
			}(),
			values: func() []string {
				values := make([]string, 1000_000)
				for i := range values {
					values[i] = fmt.Sprintf("DC%d", i)
				}
				return values
			}(),
			addrs: func() []netip.Addr {
				rng := rand.New(rand.NewSource(43))
				addrs := make([]netip.Addr, 1000)
				for i := range addrs {
					addrs[i] = netip.AddrFrom16([16]byte{
						0x20, 0x01, 0x0d, 0xb8,
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
						byte(rng.Intn(256)), byte(rng.Intn(256)),
					})
				}
				return addrs
			}(),
			isV6: true,
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			if bm.isV6 {
				// Measure memory before insertion
				runtime.GC()
				var memBefore runtime.MemStats
				runtime.ReadMemStats(&memBefore)

				// Setup: Insert 1M prefixes
				tree := string_tree.NewTreeV6()

				for i := range 1000_000 {
					addr := bm.prefixes[i].Addr()
					bits := bm.prefixes[i].Bits()
					_, _ = tree.Set(patricia.NewIPv6Address(addr.AsSlice(), uint(bits)), bm.values[i])
				}

				// Measure memory after insertion
				runtime.GC()
				var memAfter runtime.MemStats
				runtime.ReadMemStats(&memAfter)

				allocDiff := memAfter.Alloc - memBefore.Alloc
				totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc

				b.Logf("Memory usage after 1M inserts: Alloc=%d bytes (%.2f MB), TotalAlloc=%d bytes (%.2f MB)",
					allocDiff, float64(allocDiff)/(1024*1024),
					totalAllocDiff, float64(totalAllocDiff)/(1024*1024))

				b.ResetTimer()
				b.ReportAllocs()

				idx := 0
				foundCount := 0
				for b.Loop() {
					addr := bm.addrs[idx]
					ok, _ := tree.FindDeepestTag(patricia.NewIPv6Address(addr.AsSlice(), 128))
					if ok {
						foundCount++
					}
					idx = (idx + 1) % len(bm.addrs)
				}

				if foundCount == 0 {
					b.Fatalf("No successful lookups in %d iterations", b.N)
				}
			} else {
				// Measure memory before insertion
				runtime.GC()
				var memBefore runtime.MemStats
				runtime.ReadMemStats(&memBefore)

				// Setup: Insert 1M prefixes
				tree := string_tree.NewTreeV4()

				for i := range 1000_000 {
					addr := bm.prefixes[i].Addr()
					bits := bm.prefixes[i].Bits()
					_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), bm.values[i])
				}

				// Measure memory after insertion
				runtime.GC()
				var memAfter runtime.MemStats
				runtime.ReadMemStats(&memAfter)

				allocDiff := memAfter.Alloc - memBefore.Alloc
				totalAllocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc

				b.Logf("Memory usage after 1M inserts: Alloc=%d bytes (%.2f MB), TotalAlloc=%d bytes (%.2f MB)",
					allocDiff, float64(allocDiff)/(1024*1024),
					totalAllocDiff, float64(totalAllocDiff)/(1024*1024))

				b.ResetTimer()
				b.ReportAllocs()

				idx := 0
				foundCount := 0
				for b.Loop() {
					addr := bm.addrs[idx]
					ok, _ := tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
					if ok {
						foundCount++
					}
					idx = (idx + 1) % len(bm.addrs)
				}

				if foundCount == 0 {
					b.Fatalf("No successful lookups in %d iterations", b.N)
				}
			}
		})
	}
}
