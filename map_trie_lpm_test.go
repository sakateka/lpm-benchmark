package main

import (
	"fmt"
	"math/rand"
	"net/netip"
	"testing"
)

var onEmptyString = func(v string) func() string {
	return func() string { return v }
}

var onUpdateString = func(v string) func(string) string {
	return func(string) string { return v }
}

// TestMapTrieBasicOperations tests basic insert and lookup operations
func TestMapTrieBasicOperations(t *testing.T) {
	tests := []struct {
		name        string
		prefixes    []struct{ cidr, value string }
		lookups     []struct{ addr, want string }
		shouldMatch bool
	}{
		{
			name: "single IPv4 prefix",
			prefixes: []struct{ cidr, value string }{
				{"192.168.1.0/24", "DC1"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.1", "DC1"},
				{"192.168.1.255", "DC1"},
				{"192.168.2.1", ""},
			},
			shouldMatch: true,
		},
		{
			name: "overlapping IPv4 prefixes - more specific wins",
			prefixes: []struct{ cidr, value string }{
				{"10.0.0.0/8", "DC1"},
				{"10.1.0.0/16", "DC2"},
				{"10.1.1.0/24", "DC3"},
			},
			lookups: []struct{ addr, want string }{
				{"10.0.0.1", "DC1"},
				{"10.1.0.1", "DC2"},
				{"10.1.1.1", "DC3"},
				{"10.2.0.1", "DC1"},
			},
			shouldMatch: true,
		},
		{
			name: "reverse insertion order - should still work",
			prefixes: []struct{ cidr, value string }{
				{"10.1.1.0/24", "DC3"},
				{"10.1.0.0/16", "DC2"},
				{"10.0.0.0/8", "DC1"},
			},
			lookups: []struct{ addr, want string }{
				{"10.0.0.1", "DC1"},
				{"10.1.0.1", "DC2"},
				{"10.1.1.1", "DC3"},
			},
			shouldMatch: true,
		},
		{
			name: "IPv6 basic",
			prefixes: []struct{ cidr, value string }{
				{"2001:db8::/32", "DC1"},
				{"2001:db8:1::/48", "DC2"},
			},
			lookups: []struct{ addr, want string }{
				{"2001:db8::1", "DC1"},
				{"2001:db8:1::1", "DC2"},
				{"2001:db9::1", ""},
			},
			shouldMatch: true,
		},
		{
			name: "default route IPv4",
			prefixes: []struct{ cidr, value string }{
				{"0.0.0.0/0", "DEFAULT"},
				{"192.168.0.0/16", "DC1"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.1", "DC1"},
				{"8.8.8.8", "DEFAULT"},
			},
			shouldMatch: true,
		},
		{
			name: "host routes /32",
			prefixes: []struct{ cidr, value string }{
				{"192.168.1.0/24", "DC1"},
				{"192.168.1.100/32", "DC2"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.100", "DC2"},
				{"192.168.1.101", "DC1"},
			},
			shouldMatch: true,
		},
		{
			name: "adjacent prefixes",
			prefixes: []struct{ cidr, value string }{
				{"192.168.0.0/24", "DC1"},
				{"192.168.1.0/24", "DC2"},
				{"192.168.2.0/24", "DC3"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.0.1", "DC1"},
				{"192.168.1.1", "DC2"},
				{"192.168.2.1", "DC3"},
				{"192.168.3.1", ""},
			},
			shouldMatch: true,
		},
		{
			name: "non-byte-aligned prefixes",
			prefixes: []struct{ cidr, value string }{
				{"192.168.1.0/25", "DC1"},
				{"192.168.1.128/25", "DC2"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.1", "DC1"},
				{"192.168.1.127", "DC1"},
				{"192.168.1.128", "DC2"},
				{"192.168.1.255", "DC2"},
			},
			shouldMatch: true,
		},
		{
			name: "edge case - /31 prefix",
			prefixes: []struct{ cidr, value string }{
				{"192.168.1.0/31", "DC1"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.0", "DC1"},
				{"192.168.1.1", "DC1"},
				{"192.168.1.2", ""},
			},
			shouldMatch: true,
		},
		{
			name: "multiple overlapping at same byte boundary",
			prefixes: []struct{ cidr, value string }{
				{"10.0.0.0/8", "DC1"},
				{"10.10.0.0/16", "DC2"},
				{"10.10.10.0/24", "DC3"},
				{"10.10.10.10/32", "DC4"},
			},
			lookups: []struct{ addr, want string }{
				{"10.0.0.1", "DC1"},
				{"10.10.0.1", "DC2"},
				{"10.10.10.1", "DC3"},
				{"10.10.10.10", "DC4"},
			},
			shouldMatch: true,
		},
		{
			name: "IPv6 /128 host route",
			prefixes: []struct{ cidr, value string }{
				{"2001:db8::/32", "DC1"},
				{"2001:db8::1/128", "DC2"},
			},
			lookups: []struct{ addr, want string }{
				{"2001:db8::1", "DC2"},
				{"2001:db8::2", "DC1"},
			},
			shouldMatch: true,
		},
		{
			name: "duplicate value deduplication",
			prefixes: []struct{ cidr, value string }{
				{"192.168.1.0/24", "DC1"},
				{"192.168.2.0/24", "DC1"},
				{"192.168.3.0/24", "DC1"},
			},
			lookups: []struct{ addr, want string }{
				{"192.168.1.1", "DC1"},
				{"192.168.2.1", "DC1"},
				{"192.168.3.1", "DC1"},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)

			// Insert all prefixes
			for _, p := range tt.prefixes {
				prefix := netip.MustParsePrefix(p.cidr)
				trie.InsertOrUpdate(prefix, onEmptyString(p.value), onUpdateString(p.value))
			}

			// Test all lookups
			for _, l := range tt.lookups {
				addr := netip.MustParseAddr(l.addr)
				_, got, found := trie.Lookup(addr)

				if l.want == "" {
					if found {
						t.Errorf("Lookup(%s) = %q, want no match", l.addr, got)
					}
				} else {
					if !found {
						t.Errorf("Lookup(%s) = not found, want %q", l.addr, l.want)
					} else if got != l.want {
						t.Errorf("Lookup(%s) = %q, want %q", l.addr, got, l.want)
					}
				}
			}
		})
	}
}

// TestMapTrieEdgeCases tests edge cases that might break the implementation
func TestMapTrieEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MapTrie[netip.Prefix, netip.Addr, string])
		testFunc func(*testing.T, *MapTrie[netip.Prefix, netip.Addr, string])
	}{
		{
			name: "empty MapTrie lookup",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				// Don't insert anything
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				addr := netip.MustParseAddr("192.168.1.1")
				if _, val, found := trie.Lookup(addr); found {
					t.Errorf("Expected no match, got %q", val)
				}
			},
		},
		{
			name: "overwrite same prefix with different value",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				prefix := netip.MustParsePrefix("192.168.1.0/24")
				trie.InsertOrUpdate(prefix, onEmptyString("DC1"), onUpdateString("DC1"))
				trie.InsertOrUpdate(prefix, onEmptyString("DC2"), onUpdateString("DC2")) // Overwrite
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				addr := netip.MustParseAddr("192.168.1.1")
				_, val, found := trie.Lookup(addr)
				if !found {
					t.Error("Expected to find value")
				}
				// Should have the last inserted value
				if val != "DC2" {
					t.Errorf("Expected DC2, got %q", val)
				}
			},
		},
		{
			name: "many values - test value index bounds",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				addr := netip.MustParseAddr("10.100.0.1")
				// Insert 1000 values, where prefixes repeat every 256 iterations
				for i := range 1000 {
					value := fmt.Sprintf("DC%d", i)
					prefix := netip.MustParsePrefix(fmt.Sprintf("10.%d.0.0/16", i%256))
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
					// Test that lookup returns the current value after each insertion
					if i%256 == 100 {
						_, val, found := trie.Lookup(addr)
						if !found {
							panic(fmt.Sprintf("Expected to find value after inserting DC%d", i))
						}
						expectedValue := fmt.Sprintf("DC%d", i)
						if val != expectedValue {
							panic(fmt.Sprintf("After inserting DC%d, expected %s, got %q", i, expectedValue, val))
						}
					}
				}
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				addr := netip.MustParseAddr("10.100.0.1")
				_, val, found := trie.Lookup(addr)
				if !found {
					t.Error("Expected to find value")
				}
				if val != "DC868" {
					t.Errorf("Expected DC868 (last inserted value for 10.100.0.0/16), got %q", val)
				}
			},
		},
		{
			name: "all 256 values in first byte",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				// Insert a /8 for every possible first byte
				for i := range 256 {
					value := fmt.Sprintf("DC%d", i)
					prefix := netip.MustParsePrefix(fmt.Sprintf("%d.0.0.0/8", i))
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
				}
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				for i := range 256 {
					addr := netip.MustParseAddr(fmt.Sprintf("%d.0.0.1", i))
					_, val, found := trie.Lookup(addr)
					want := fmt.Sprintf("DC%d", i)
					if !found {
						t.Errorf("Lookup(%s) not found", addr)
					} else if val != want {
						t.Errorf("Lookup(%s) = %q, want %q", addr, val, want)
					}
				}
			},
		},
		{
			name: "deeply nested prefixes",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				// Create a deep nesting: /8, /16, /24, /32
				prefixes := []string{
					"10.0.0.0/8",
					"10.1.0.0/16",
					"10.1.1.0/24",
					"10.1.1.1/32",
				}
				for i, p := range prefixes {
					value := fmt.Sprintf("DC%d", i)
					prefix := netip.MustParsePrefix(p)
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
				}
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				tests := []struct {
					addr string
					want string
				}{
					{"10.0.0.1", "DC0"},
					{"10.1.0.1", "DC1"},
					{"10.1.1.1", "DC3"},
					{"10.1.1.2", "DC2"},
				}
				for _, tt := range tests {
					addr := netip.MustParseAddr(tt.addr)
					_, val, found := trie.Lookup(addr)
					if !found {
						t.Errorf("Lookup(%s) not found", tt.addr)
					} else if val != tt.want {
						t.Errorf("Lookup(%s) = %q, want %q", tt.addr, val, tt.want)
					}
				}
			},
		},
		{
			name: "IPv6 with many blocks",
			setup: func(trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				// Insert multiple IPv6 prefixes
				for i := range 100 {
					value := fmt.Sprintf("DC%d", i)
					prefix := netip.MustParsePrefix(fmt.Sprintf("2001:db8:%x::/48", i))
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
				}
			},
			testFunc: func(t *testing.T, trie *MapTrie[netip.Prefix, netip.Addr, string]) {
				addr := netip.MustParseAddr("2001:db8:50::1")
				_, val, found := trie.Lookup(addr)
				if !found {
					t.Error("Expected to find value")
				}
				if val != "DC80" {
					t.Errorf("Expected DC80, got %q", val)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)
			tt.setup(&trie)
			tt.testFunc(t, &trie)
		})
	}
}

// TestMapTrieLongestPrefixMatch verifies that longest prefix matching works correctly
func TestMapTrieLongestPrefixMatch(t *testing.T) {
	trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)

	// Insert prefixes in random order
	prefixes := []struct {
		cidr  string
		value string
	}{
		{"10.0.0.0/8", "DC1"},
		{"10.1.0.0/16", "DC2"},
		{"10.1.1.0/24", "DC3"},
		{"10.1.1.128/25", "DC4"},
		{"10.1.1.192/26", "DC5"},
		{"10.1.1.224/27", "DC6"},
		{"10.1.1.240/28", "DC7"},
		{"10.1.1.248/29", "DC8"},
		{"10.1.1.252/30", "DC9"},
		{"10.1.1.254/31", "DC10"},
	}

	for _, p := range prefixes {
		prefix := netip.MustParsePrefix(p.cidr)
		trie.InsertOrUpdate(prefix, onEmptyString(p.value), onUpdateString(p.value))
	}

	tests := []struct {
		addr string
		want string
	}{
		{"10.0.0.1", "DC1"},
		{"10.1.0.1", "DC2"},
		{"10.1.1.1", "DC3"},
		{"10.1.1.129", "DC4"},
		{"10.1.1.193", "DC5"},
		{"10.1.1.225", "DC6"},
		{"10.1.1.241", "DC7"},
		{"10.1.1.249", "DC8"},
		{"10.1.1.253", "DC9"},
		{"10.1.1.254", "DC10"},
		{"10.1.1.255", "DC10"},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			addr := netip.MustParseAddr(tt.addr)
			_, got, found := trie.Lookup(addr)
			if !found {
				t.Errorf("Lookup(%s) not found", tt.addr)
			} else if got != tt.want {
				t.Errorf("Lookup(%s) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

// BenchmarkMapTrieInsert benchmarks insertion performance
func BenchmarkMapTrieInsert(b *testing.B) {
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
				trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)
				for j, cidr := range bm.prefixes {
					value := fmt.Sprintf("DC%d", j)
					prefix := netip.MustParsePrefix(cidr)
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
				}
			}
		})
	}
}

// BenchmarkMapTrieLookup benchmarks lookup performance
func BenchmarkMapTrieLookup(b *testing.B) {
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
			trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)
			for j, cidr := range bm.prefixes {
				value := fmt.Sprintf("DC%d", j)
				prefix := netip.MustParsePrefix(cidr)
				trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
			}

			addrs := make([]netip.Addr, len(bm.lookups))
			for i, lookup := range bm.lookups {
				addrs[i] = netip.MustParseAddr(lookup)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for b.Loop() {
				for _, addr := range addrs {
					_, _, _ = trie.Lookup(addr)
				}
			}
		})
	}
}

// BenchmarkMapTrieInsertAndLookup benchmarks combined insert and lookup
func BenchmarkMapTrieInsertAndLookup(b *testing.B) {
	prefixes := make([]string, 1000)
	for i := range 1000 {
		prefixes[i] = fmt.Sprintf("10.%d.%d.0/24", i/256, i%256)
	}

	b.ReportAllocs()

	for b.Loop() {
		trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)

		// Insert
		for j, cidr := range prefixes {
			value := fmt.Sprintf("DC%d", j)
			prefix := netip.MustParsePrefix(cidr)
			trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
		}

		// Lookup
		for j := range 100 {
			addr := netip.MustParseAddr(fmt.Sprintf("10.%d.%d.1", j/256, j%256))
			_, _, _ = trie.Lookup(addr)
		}
	}
}

// BenchmarkMapTrieMemoryFootprint measures memory usage
func BenchmarkMapTrieMemoryFootprint(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("prefixes_%d", size), func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)

				for j := range size {
					value := fmt.Sprintf("DC%d", j)
					prefix := netip.MustParsePrefix(
						fmt.Sprintf("10.%d.%d.0/24", j/256, j%256))
					trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
				}

				// Force allocation tracking
				_ = trie.Len()
			}
		})
	}
}

// BenchmarkMapTrieConcurrentLookup benchmarks concurrent lookups
func BenchmarkMapTrieConcurrentLookup(b *testing.B) {
	trie := NewMapTrie[netip.Prefix, netip.Addr, string](0)

	// Setup with 100 prefixes
	for i := range 100 {
		value := fmt.Sprintf("DC%d", i)
		prefix := netip.MustParsePrefix(fmt.Sprintf("10.%d.0.0/16", i%256))
		trie.InsertOrUpdate(prefix, onEmptyString(value), onUpdateString(value))
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
			_, _, _ = trie.Lookup(addr)
		}
	})
}
