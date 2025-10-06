```txt
goos: linux
goarch: amd64
pkg: github.com/sakateka/lpm-benchmark
cpu: 13th Gen Intel(R) Core(TM) i7-13700H
BenchmarkLPMInsert1M/ipv4_1M_prefixes-20         	 1443982	       900.6 ns/op	     378 B/op	       0 allocs/op
BenchmarkLPMInsert1M/ipv6_1M_prefixes-20         	 1000000	      1373 ns/op	    3379 B/op	       2 allocs/op
BenchmarkLPMLookup1M/ipv4_1M_prefixes-20         	224538270	         5.346 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkLPMLookup1M/ipv4_1M_prefixes-20
    lpm_bench1M_test.go:206: Memory usage after 1M inserts: Alloc=408613656 bytes (389.68 MB), TotalAlloc=546131048 bytes (520.83 MB)
    lpm_bench1M_test.go:210: lpm.v4StorageSize: 334705488, lpm.v6StorageSize: 1056
    lpm_bench1M_test.go:211: lpm.v4Blocks: 324327, lpm.v6Blocks: 1, total size: 394595522
BenchmarkLPMLookup1M/ipv6_1M_prefixes-20         	122018863	         9.842 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkLPMLookup1M/ipv6_1M_prefixes-20
    lpm_bench1M_test.go:206: Memory usage after 1M inserts: Alloc=3142905768 bytes (2997.31 MB), TotalAlloc=3379357976 bytes (3222.81 MB)
    lpm_bench1M_test.go:210: lpm.v4StorageSize: 1056, lpm.v6StorageSize: 3065450760
    lpm_bench1M_test.go:211: lpm.v4Blocks: 1, lpm.v6Blocks: 2970398, total size: 3125340794
BenchmarkLPMInsert/single_prefix-20              	 1317882	       920.1 ns/op	    4615 B/op	      13 allocs/op
BenchmarkLPMInsert/10_prefixes-20                	  257604	      4376 ns/op	   16023 B/op	      41 allocs/op
BenchmarkLPMInsert/100_prefixes-20               	   83466	     14486 ns/op	   15101 B/op	     126 allocs/op
BenchmarkLPMInsert/overlapping_prefixes-20       	  449498	      2640 ns/op	    6606 B/op	      30 allocs/op
BenchmarkLPMInsert/ipv6_prefixes-20              	  451050	      2513 ns/op	    9910 B/op	      24 allocs/op
BenchmarkLPMLookup/single_prefix_match-20        	188756667	         6.473 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMLookup/10_prefixes_various_matches-20         	38302705	        31.92 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMLookup/100_prefixes_deep_lookup-20            	69333266	        16.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMLookup/no_match-20                            	412885456	         2.910 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMLookup/longest_prefix_match-20                	163955060	         7.264 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMLookup/ipv6_lookup-20                         	129085310	         9.295 ns/op	       0 B/op	       0 allocs/op
BenchmarkLPMInsertAndLookup-20                            	    6234	    163127 ns/op	  166477 B/op	    1890 allocs/op
BenchmarkLPMMemoryFootprint/prefixes_10-20                	  401997	      3172 ns/op	    5709 B/op	      39 allocs/op
BenchmarkLPMMemoryFootprint/prefixes_100-20               	   55663	     21694 ns/op	   17366 B/op	     228 allocs/op
BenchmarkLPMMemoryFootprint/prefixes_1000-20              	    5166	    233058 ns/op	  180566 B/op	    2790 allocs/op
BenchmarkLPMMemoryFootprint/prefixes_10000-20             	     513	   2354657 ns/op	 1903348 B/op	   29897 allocs/op
BenchmarkLPMConcurrentLookup-20                           	1000000000	         0.8724 ns/op	       0 B/op	       0 allocs/op

BenchmarkMapTrieInsert1M/ipv4_1M_prefixes-20              	10594519	       111.1 ns/op	      57 B/op	       2 allocs/op
BenchmarkMapTrieInsert1M/ipv6_1M_prefixes-20              	 9330356	       131.6 ns/op	      62 B/op	       2 allocs/op
BenchmarkMapTrieLookup1M/ipv4_1M_prefixes-20              	 1617556	       741.5 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkMapTrieLookup1M/ipv4_1M_prefixes-20
    map_trie_bench1M_test.go:206: Memory usage after 1M inserts: Alloc=52852080 bytes (50.40 MB), TotalAlloc=153625288 bytes (146.51 MB)
BenchmarkMapTrieLookup1M/ipv6_1M_prefixes-20              	  372284	      3216 ns/op	       0 B/op	       0 allocs/op
--- BENCH: BenchmarkMapTrieLookup1M/ipv6_1M_prefixes-20
    map_trie_bench1M_test.go:206: Memory usage after 1M inserts: Alloc=66401848 bytes (63.33 MB), TotalAlloc=180520496 bytes (172.16 MB)
BenchmarkMapTrieInsert/single_prefix-20                   	  350397	      3226 ns/op	    6665 B/op	     133 allocs/op
BenchmarkMapTrieInsert/10_prefixes-20                     	  246210	      4772 ns/op	    8375 B/op	     163 allocs/op
BenchmarkMapTrieInsert/100_prefixes-20                    	   61016	     19592 ns/op	   24373 B/op	     439 allocs/op
BenchmarkMapTrieInsert/overlapping_prefixes-20            	  256677	      4623 ns/op	    8375 B/op	     163 allocs/op
BenchmarkMapTrieInsert/ipv6_prefixes-20                   	  329271	      3809 ns/op	    7651 B/op	     144 allocs/op
BenchmarkMapTrieLookup/single_prefix_match-20             	28546726	        41.71 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieLookup/10_prefixes_various_matches-20     	 2353669	       511.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieLookup/100_prefixes_deep_lookup-20        	 5319297	       225.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieLookup/no_match-20                        	 8313399	       140.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieLookup/longest_prefix_match-20            	31410326	        40.89 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieLookup/ipv6_lookup-20                     	 3424334	       342.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkMapTrieInsertAndLookup-20                        	    5271	    222288 ns/op	  295120 B/op	    3994 allocs/op
BenchmarkMapTrieMemoryFootprint/prefixes_10-20            	  206577	      5623 ns/op	    8191 B/op	     173 allocs/op
BenchmarkMapTrieMemoryFootprint/prefixes_100-20           	   43488	     27581 ns/op	   25581 B/op	     539 allocs/op
BenchmarkMapTrieMemoryFootprint/prefixes_1000-20          	    4015	    280725 ns/op	  309194 B/op	    4894 allocs/op
BenchmarkMapTrieMemoryFootprint/prefixes_10000-20         	     411	   2897015 ns/op	 2638543 B/op	   49956 allocs/op
BenchmarkMapTrieConcurrentLookup-20                       	120090297	        10.09 ns/op	       0 B/op	       0 allocs/op

BenchmarkPatriciaInsert1M/ipv4_1M_prefixes-20             	 8459772	       119.6 ns/op	      20 B/op	       1 allocs/op
BenchmarkPatriciaInsert1M/ipv6_1M_prefixes-20             	 5338027	       195.4 ns/op	      64 B/op	       1 allocs/op
BenchmarkPatriciaLookup1M/ipv4_1M_prefixes-20             	61557802	        20.93 ns/op	       4 B/op	       1 allocs/op
--- BENCH: BenchmarkPatriciaLookup1M/ipv4_1M_prefixes-20
    patricia_bench1M_test.go:274: Memory usage after 1M inserts: Alloc=69903664 bytes (66.67 MB), TotalAlloc=139806904 bytes (133.33 MB)
BenchmarkPatriciaLookup1M/ipv6_1M_prefixes-20             	37964596	        31.72 ns/op	      16 B/op	       1 allocs/op
--- BENCH: BenchmarkPatriciaLookup1M/ipv6_1M_prefixes-20
    patricia_bench1M_test.go:230: Memory usage after 1M inserts: Alloc=128623920 bytes (122.67 MB), TotalAlloc=257246968 bytes (245.33 MB)
BenchmarkPatriciaInsert/single_prefix_v4-20               	 4333126	       275.3 ns/op	     648 B/op	       7 allocs/op
BenchmarkPatriciaInsert/10_prefixes_v4-20                 	  687667	      1742 ns/op	    3034 B/op	      30 allocs/op
BenchmarkPatriciaInsert/100_prefixes_v4-20                	   69078	     17203 ns/op	   28322 B/op	     219 allocs/op
BenchmarkPatriciaInsert/overlapping_prefixes_v4-20        	  671061	      1782 ns/op	    3034 B/op	      30 allocs/op
BenchmarkPatriciaInsert/ipv6_prefixes_v6-20               	 1000000	      1060 ns/op	    3023 B/op	      15 allocs/op
BenchmarkPatriciaLookup/single_prefix_match_v4-20         	92596279	        12.02 ns/op	       4 B/op	       1 allocs/op
BenchmarkPatriciaLookup/10_prefixes_various_matches_v4-20 	11666076	       101.9 ns/op	      24 B/op	       6 allocs/op
BenchmarkPatriciaLookup/100_prefixes_deep_lookup_v4-20    	16007851	        78.63 ns/op	      12 B/op	       3 allocs/op
BenchmarkPatriciaLookup/no_match_v4-20                    	128040733	         9.354 ns/op	       4 B/op	       1 allocs/op
BenchmarkPatriciaLookup/longest_prefix_match_v4-20        	51200364	        23.21 ns/op	       4 B/op	       1 allocs/op
BenchmarkPatriciaLookup/ipv6_lookup_v6-20                 	53704759	        21.91 ns/op	      16 B/op	       1 allocs/op
BenchmarkPatriciaInsertAndLookup-20                       	    5534	    217779 ns/op	  295937 B/op	    2977 allocs/op
BenchmarkPatriciaMemoryFootprint/prefixes_10-20           	  456146	      2587 ns/op	    3168 B/op	      40 allocs/op
BenchmarkPatriciaMemoryFootprint/prefixes_100-20          	   47433	     25060 ns/op	   29655 B/op	     319 allocs/op
BenchmarkPatriciaMemoryFootprint/prefixes_1000-20         	    4285	    270824 ns/op	  307684 B/op	    3777 allocs/op
BenchmarkPatriciaMemoryFootprint/prefixes_10000-20        	     411	   2878436 ns/op	 3885180 B/op	   39843 allocs/op
BenchmarkPatriciaConcurrentLookup-20                      	159143098	         7.442 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/sakateka/lpm-benchmark	81.103s
```


Pytricia results
```
PyTricia IPV4 benchmark
  prefixes:         1,000,000
  insert:           0.133061 s  |  7,515,372.42 ops/s  |  133.06 ns/op
  lookup probes:    5,000,000
  lookup:           2.245650 s  |  2,226,527.29 ops/s  |  449.13 ns/op  |  found=350,000
  RSS after insert: 291,655,680 B (278.14 MB)
  RSS delta:        269,025,280 B (256.56 MB)
PyTricia IPV6 benchmark
  prefixes:         1,000,000
  insert:           0.267654 s  |  3,736,165.93 ops/s  |  267.65 ns/op
  lookup probes:    5,000,000
  lookup:           2.779728 s  |  1,798,737.35 ops/s  |  555.95 ns/op  |  found=5,000,000
  RSS after insert: 361,086,976 B (344.36 MB)
  RSS delta:        304,087,040 B (290.00 MB)
```
