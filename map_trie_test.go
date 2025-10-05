package main

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var onEmpty = func(v int) func() int {
	return func() int { return v }
}

var onUpdate = func(v int) func(int) int {
	return func(int) int { return v }
}

func Test_MapTrie_LookupEmpty(t *testing.T) {
	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)

	// Expect failed lookup in empty trie.
	_, _, ok := trie.Lookup(netip.MustParseAddr("192.168.9.1"))
	assert.False(t, ok)
}

func Test_MapTrie_LookupAfterInsert(t *testing.T) {
	cases := []struct {
		addr           string
		expectedOk     bool
		expectedPrefix netip.Prefix
		expectedIdx    int
	}{
		{"192.168.9.1", true, netip.MustParsePrefix("192.168.0.0/16"), 0},
		{"127.0.0.1", false, netip.Prefix{}, 0},
	}

	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.0.0/16"), onEmpty(0), onUpdate(0))

	for _, c := range cases {
		prefix, v, ok := trie.Lookup(netip.MustParseAddr(c.addr))
		require.Equal(t, c.expectedOk, ok)
		assert.Equal(t, c.expectedIdx, v)
		assert.Equal(t, c.expectedPrefix, prefix)
	}
}

func Test_MapTrie_LookupAfterInsertUpdate(t *testing.T) {
	cases := []struct {
		addr           string
		expectedOk     bool
		expectedPrefix netip.Prefix
		expectedIdx    int
	}{
		{"192.168.9.1", true, netip.MustParsePrefix("192.168.0.0/16"), 1},
		{"127.0.0.1", false, netip.Prefix{}, 0},
	}

	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.0.0/16"), onEmpty(0), onUpdate(0))
	// This should update the value to 1.
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.0.0/16"), onEmpty(1), onUpdate(1))

	for _, c := range cases {
		prefix, v, ok := trie.Lookup(netip.MustParseAddr(c.addr))
		require.Equal(t, c.expectedOk, ok)
		assert.Equal(t, c.expectedIdx, v)
		assert.Equal(t, c.expectedPrefix, prefix)
	}
}

func Test_MapTrie_LookupAfterInsertNestedPrefixes(t *testing.T) {
	cases := []struct {
		addr           string
		expectedOk     bool
		expectedPrefix netip.Prefix
		expectedIdx    int
	}{
		{"192.168.1.1", true, netip.MustParsePrefix("192.168.1.1/32"), 4},
		{"192.168.1.2", true, netip.MustParsePrefix("192.168.1.0/24"), 3},
		{"192.168.2.2", true, netip.MustParsePrefix("192.168.0.0/16"), 2},
		{"192.200.1.1", true, netip.MustParsePrefix("192.0.0.0/8"), 1},
		{"127.0.0.1", true, netip.MustParsePrefix("0.0.0.0/0"), 0},
	}

	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)
	trie.InsertOrUpdate(netip.MustParsePrefix("0.0.0.0/0"), onEmpty(0), onUpdate(0))
	trie.InsertOrUpdate(netip.MustParsePrefix("192.0.0.0/8"), onEmpty(1), onUpdate(1))
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.0.0/16"), onEmpty(2), onUpdate(2))
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.1.0/24"), onEmpty(3), onUpdate(3))
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.1.1/32"), onEmpty(4), onUpdate(4))

	for _, c := range cases {
		prefix, v, ok := trie.Lookup(netip.MustParseAddr(c.addr))
		require.Equal(t, c.expectedOk, ok)
		assert.Equal(t, c.expectedIdx, v)
		assert.Equal(t, c.expectedPrefix, prefix)
	}
}

func Test_MapTrie_Lookup6(t *testing.T) {
	cases := []struct {
		prefix         string
		expectedOk     bool
		expectedPrefix netip.Prefix
		expectedIdx    int
	}{
		{"fd25:cf19:6b13:cafe:babe:be57:f00d:04a5/128", false, netip.Prefix{}, 0},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d:400/120", false, netip.Prefix{}, 0},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d::/112", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 2},
		{"fd25:cf19:6b13:cafe::/64", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 2},
		{"fd25:8888:6b13:cafe::/64", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 2},
		{"fd25::/16", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 2},
	}

	addr := netip.MustParseAddr("fd25:cf19:6b13:cafe:babe:be57:f00d:0001")

	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)

	for idx, c := range cases {
		prefix := netip.MustParsePrefix(c.prefix).Masked()

		trie.InsertOrUpdate(prefix, onEmpty(idx), onUpdate(idx))

		matchedPrefix, value, ok := trie.Lookup(addr)
		require.Equal(t, c.expectedOk, ok,
			"lookup expected match==%t, but ok=%t, prefix=%s", c.expectedOk, ok, prefix)
		require.Equal(t, c.expectedIdx, value,
			"lookup expected value==%d, but value=%d, prefix=%s", c.expectedIdx, value, prefix)
		require.Equal(t, c.expectedPrefix, matchedPrefix)
	}
}

func Test_MapTrie_Lookup6TopDownInsert(t *testing.T) {
	cases := []struct {
		prefix         string
		expectedOk     bool
		expectedPrefix netip.Prefix
		expectedIdx    int
	}{
		{"fd25::/16", true, netip.MustParsePrefix("fd25::/16"), 0},
		{"fd25:8888:6b13:cafe::/64", true, netip.MustParsePrefix("fd25::/16"), 0},
		{"fd25:cf19:6b13:cafe::/64", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe::/64"), 2},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d::/112", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 3},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d:400/120", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 3},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d:04a5/128", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d::/112"), 3},
		{"fd25:cf19:6b13:cafe:babe:be57:f00d:0001/128", true, netip.MustParsePrefix("fd25:cf19:6b13:cafe:babe:be57:f00d:0001/128"), 6},
	}

	addr := netip.MustParseAddr("fd25:cf19:6b13:cafe:babe:be57:f00d:0001")

	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)

	for idx, c := range cases {
		prefix := netip.MustParsePrefix(c.prefix).Masked()

		trie.InsertOrUpdate(prefix, onEmpty(idx), onUpdate(idx))

		matchedPrefix, value, ok := trie.Lookup(addr)
		require.Equal(t, c.expectedOk, ok,
			"lookup expected match==%t, but ok=%t, prefix=%s", c.expectedOk, ok, prefix)
		require.Equal(t, c.expectedIdx, value,
			"lookup expected value==%d, but value=%d, prefix=%s", c.expectedIdx, value, prefix)
		require.Equal(t, c.expectedPrefix, matchedPrefix)
	}
}

func Test_MapTrie_LookupTraverse(t *testing.T) {
	trie := NewMapTrie[netip.Prefix, netip.Addr, int](0)

	traverseLPM := func(addr netip.Addr) []netip.Prefix {
		out := make([]netip.Prefix, 0)
		trie.LookupTraverse(addr, func(prefix netip.Prefix, value int) bool {
			out = append(out, prefix)
			return true
		})

		return out
	}

	addr := netip.MustParseAddr("192.168.9.32")
	assert.Equal(t, []netip.Prefix{}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.9.32/32"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.9.0/24"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// Note, that 192.168.9.0/27 does not contain 192.168.9.32 ...
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.9.0/27"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// ... but 192.168.9.0/26 does.
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.9.0/26"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// Does not affect.
	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.10.0/24"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("192.168.0.0/16"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// 192.168.0.0 in hex.
	trie.InsertOrUpdate(netip.MustParsePrefix("a8c0::/16"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("a8c0::/112"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// v6 mapped ::ffff:168.192.1.9
	trie.InsertOrUpdate(netip.MustParsePrefix("::ffff:a8c0:109/16"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// v6 mapped ::ffff:168.192.1.9
	trie.InsertOrUpdate(netip.MustParsePrefix("::ffff:a8c0:109/112"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("192.0.0.0/8"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	trie.InsertOrUpdate(netip.MustParsePrefix("193.168.9.1/8"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// NOTE: this is very important case! No intermix between IPv4 and IPv6 ...
	trie.InsertOrUpdate(netip.MustParsePrefix("::/0"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("192.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))

	// ... but IPv4 UNSPECIFIED is okay.
	trie.InsertOrUpdate(netip.MustParsePrefix("0.0.0.0/0"), onEmpty(0), onUpdate(0))
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/0"),
		netip.MustParsePrefix("192.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.9.0/24"),
		netip.MustParsePrefix("192.168.9.0/26"),
		netip.MustParsePrefix("192.168.9.32/32"),
	}, traverseLPM(addr))
}
