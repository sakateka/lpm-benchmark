package main

import (
	"net/netip"
	"testing"

	"github.com/kentik/patricia"
	"github.com/kentik/patricia/string_tree"
)

// TestPatriciaSmallerThenLargerRange tests the scenario where:
// 1. A smaller range is inserted first with value X
// 2. A larger range that includes the smaller range is inserted with value Y
// 3. Addresses after the smaller range (but still in the larger range) should return Y
func TestPatriciaSmallerThenLargerRange(t *testing.T) {
	tests := []struct {
		name    string
		inserts []struct{ cidr, value string }
		lookups []struct{ addr, want string }
	}{
		{
			name: "smaller /24 then larger /16",
			inserts: []struct{ cidr, value string }{
				{"10.1.1.0/24", "SMALL"}, // Insert smaller range first
				{"10.1.0.0/16", "LARGE"}, // Then insert larger range that includes it
			},
			lookups: []struct{ addr, want string }{
				// Addresses in the smaller range should still return SMALL (more specific)
				{"10.1.1.1", "SMALL"},
				{"10.1.1.100", "SMALL"},
				{"10.1.1.255", "SMALL"},

				// Addresses AFTER the smaller range but still in the larger range
				// should return LARGE
				{"10.1.2.1", "LARGE"},
				{"10.1.3.1", "LARGE"},
				{"10.1.255.1", "LARGE"},

				// Addresses BEFORE the smaller range but in the larger range
				{"10.1.0.1", "LARGE"},
			},
		},
		{
			name: "smaller /25 then larger /24",
			inserts: []struct{ cidr, value string }{
				{"192.168.1.0/25", "SMALL"}, // 192.168.1.0 - 192.168.1.127
				{"192.168.1.0/24", "LARGE"}, // 192.168.1.0 - 192.168.1.255
			},
			lookups: []struct{ addr, want string }{
				// In the smaller range
				{"192.168.1.1", "SMALL"},
				{"192.168.1.127", "SMALL"},

				// After the smaller range, should match larger range
				{"192.168.1.128", "LARGE"},
				{"192.168.1.200", "LARGE"},
				{"192.168.1.255", "LARGE"},
			},
		},
		{
			name: "multiple smaller ranges then larger",
			inserts: []struct{ cidr, value string }{
				{"10.0.1.0/24", "SMALL1"},
				{"10.0.3.0/24", "SMALL2"},
				{"10.0.5.0/24", "SMALL3"},
				{"10.0.0.0/16", "LARGE"}, // Should cover all gaps
			},
			lookups: []struct{ addr, want string }{
				// Specific ranges
				{"10.0.1.1", "SMALL1"},
				{"10.0.3.1", "SMALL2"},
				{"10.0.5.1", "SMALL3"},

				// Gaps between specific ranges - should match LARGE
				{"10.0.0.1", "LARGE"},
				{"10.0.2.1", "LARGE"}, // Between SMALL1 and SMALL2
				{"10.0.4.1", "LARGE"}, // Between SMALL2 and SMALL3
				{"10.0.6.1", "LARGE"}, // After SMALL3
				{"10.0.255.1", "LARGE"},
			},
		},
		{
			name: "smaller /32 then larger /24",
			inserts: []struct{ cidr, value string }{
				{"172.16.1.100/32", "HOST"},
				{"172.16.1.0/24", "SUBNET"},
			},
			lookups: []struct{ addr, want string }{
				{"172.16.1.100", "HOST"},
				{"172.16.1.1", "SUBNET"},
				{"172.16.1.99", "SUBNET"},
				{"172.16.1.101", "SUBNET"}, // Right after the host
				{"172.16.1.255", "SUBNET"},
			},
		},
		{
			name: "non-byte-aligned smaller then larger",
			inserts: []struct{ cidr, value string }{
				{"10.1.1.64/26", "SMALL"}, // 10.1.1.64 - 10.1.1.127
				{"10.1.1.0/24", "LARGE"},  // 10.1.1.0 - 10.1.1.255
			},
			lookups: []struct{ addr, want string }{
				// Before smaller range
				{"10.1.1.1", "LARGE"},
				{"10.1.1.63", "LARGE"},

				// In smaller range
				{"10.1.1.64", "SMALL"},
				{"10.1.1.100", "SMALL"},
				{"10.1.1.127", "SMALL"},

				// After smaller range
				{"10.1.1.128", "LARGE"},
				{"10.1.1.200", "LARGE"},
				{"10.1.1.255", "LARGE"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := string_tree.NewTreeV4()

			// Insert all prefixes in order
			for _, ins := range tt.inserts {
				prefix := netip.MustParsePrefix(ins.cidr)
				addr := prefix.Addr()
				bits := prefix.Bits()
				_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), ins.value)
			}

			// Test all lookups
			for _, l := range tt.lookups {
				addr := netip.MustParseAddr(l.addr)
				found, tag := tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))

				if !found {
					t.Errorf("Lookup(%s) = not found, want %q", l.addr, l.want)
				} else if tag != l.want {
					t.Errorf("Lookup(%s) = %q, want %q", l.addr, tag, l.want)
				}
			}
		})
	}
}

// TestPatriciaReverseInsertionOrder tests that insertion order shouldn't matter
func TestPatriciaReverseInsertionOrder(t *testing.T) {
	t.Run("larger then smaller - should work", func(t *testing.T) {
		tree := string_tree.NewTreeV4()

		// Insert larger range first
		prefix := netip.MustParsePrefix("10.1.0.0/16")
		addr := prefix.Addr()
		bits := prefix.Bits()
		_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), "LARGE")

		// Then insert smaller range
		prefix = netip.MustParsePrefix("10.1.1.0/24")
		addr = prefix.Addr()
		bits = prefix.Bits()
		_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), "SMALL")

		// Test lookups
		tests := []struct{ addr, want string }{
			{"10.1.0.1", "LARGE"},
			{"10.1.1.1", "SMALL"},
			{"10.1.2.1", "LARGE"},
		}

		for _, tt := range tests {
			addr := netip.MustParseAddr(tt.addr)
			found, tag := tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
			if !found || tag != tt.want {
				t.Errorf("Lookup(%s) = %q (found=%v), want %q", tt.addr, tag, found, tt.want)
			}
		}
	})

	t.Run("smaller then larger - should also work", func(t *testing.T) {
		tree := string_tree.NewTreeV4()

		// Insert smaller range first
		prefix := netip.MustParsePrefix("10.1.1.0/24")
		addr := prefix.Addr()
		bits := prefix.Bits()
		_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), "SMALL")

		// Then insert larger range
		prefix = netip.MustParsePrefix("10.1.0.0/16")
		addr = prefix.Addr()
		bits = prefix.Bits()
		_, _ = tree.Set(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), uint(bits)), "LARGE")

		// Test lookups - these should give the same results as above
		tests := []struct{ addr, want string }{
			{"10.1.0.1", "LARGE"},
			{"10.1.1.1", "SMALL"}, // More specific should win
			{"10.1.2.1", "LARGE"},
		}

		for _, tt := range tests {
			addr := netip.MustParseAddr(tt.addr)
			found, tag := tree.FindDeepestTag(patricia.NewIPv4AddressFromBytes(addr.AsSlice(), 32))
			if !found || tag != tt.want {
				t.Errorf("Lookup(%s) = %q (found=%v), want %q", tt.addr, tag, found, tt.want)
			}
		}
	})
}

// TestPatriciaIPv6SmallerThenLargerRange tests the same scenarios for IPv6
func TestPatriciaIPv6SmallerThenLargerRange(t *testing.T) {
	tests := []struct {
		name    string
		inserts []struct{ cidr, value string }
		lookups []struct{ addr, want string }
	}{
		{
			name: "smaller /48 then larger /32",
			inserts: []struct{ cidr, value string }{
				{"2001:db8:1::/48", "SMALL"},
				{"2001:db8::/32", "LARGE"},
			},
			lookups: []struct{ addr, want string }{
				{"2001:db8:1::1", "SMALL"},
				{"2001:db8:2::1", "LARGE"},
				{"2001:db8::1", "LARGE"},
			},
		},
		{
			name: "smaller /128 then larger /64",
			inserts: []struct{ cidr, value string }{
				{"2001:db8::1/128", "HOST"},
				{"2001:db8::/64", "SUBNET"},
			},
			lookups: []struct{ addr, want string }{
				{"2001:db8::1", "HOST"},
				{"2001:db8::2", "SUBNET"},
				{"2001:db8::ffff", "SUBNET"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := string_tree.NewTreeV6()

			// Insert all prefixes in order
			for _, ins := range tt.inserts {
				prefix := netip.MustParsePrefix(ins.cidr)
				addr := prefix.Addr()
				bits := prefix.Bits()
				_, _ = tree.Set(patricia.NewIPv6Address(addr.AsSlice(), uint(bits)), ins.value)
			}

			// Test all lookups
			for _, l := range tt.lookups {
				addr := netip.MustParseAddr(l.addr)
				found, tag := tree.FindDeepestTag(patricia.NewIPv6Address(addr.AsSlice(), 128))

				if !found {
					t.Errorf("Lookup(%s) = not found, want %q", l.addr, l.want)
				} else if tag != l.want {
					t.Errorf("Lookup(%s) = %q, want %q", l.addr, tag, l.want)
				}
			}
		})
	}
}
