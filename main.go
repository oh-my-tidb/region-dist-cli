package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/kvproto/pkg/pdpb"
	"github.com/pingcap/pd/table"
)

var pd = flag.String("pd", "http://127.0.0.1:2379", "pd address")

type RegionInfo struct {
	ID          uint64              `json:"id"`
	StartKey    string              `json:"start_key"`
	EndKey      string              `json:"end_key"`
	RegionEpoch *metapb.RegionEpoch `json:"epoch,omitempty"`
	Peers       []*metapb.Peer      `json:"peers,omitempty"`

	Leader          *metapb.Peer      `json:"leader,omitempty"`
	DownPeers       []*pdpb.PeerStats `json:"down_peers,omitempty"`
	PendingPeers    []*metapb.Peer    `json:"pending_peers,omitempty"`
	WrittenBytes    uint64            `json:"written_bytes,omitempty"`
	ReadBytes       uint64            `json:"read_bytes,omitempty"`
	WrittenKeys     uint64            `json:"written_keys,omitempty"`
	ReadKeys        uint64            `json:"read_keys,omitempty"`
	ApproximateSize int64             `json:"approximate_size,omitempty"`
	ApproximateKeys int64             `json:"approximate_keys,omitempty"`
}

type RegionsInfo struct {
	Count   int           `json:"count"`
	Regions []*RegionInfo `json:"regions"`
}

type StoreInfo struct {
	Store struct {
		ID uint64 `json:"id"`
	} `json:"store"`
}

type StoresInfo struct {
	Stores []*StoreInfo `json:"stores"`
}

func main() {
	res, err := http.Get(*pd + "/pd/api/v1/stores")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	var stores StoresInfo
	err = json.NewDecoder(res.Body).Decode(&stores)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err = http.Get(*pd + "/pd/api/v1/regions")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	var regions RegionsInfo
	err = json.NewDecoder(res.Body).Decode(&regions)
	if err != nil {
		fmt.Println(err)
		return
	}
	print(stores.Stores, regions.Regions)
}

func print(stores []*StoreInfo, regions []*RegionInfo) {
	sort.Slice(regions, func(i, j int) bool { return regions[i].StartKey < regions[j].StartKey })
	maxKeyLen := 6
	for _, r := range regions {
		r.StartKey, r.EndKey = convertKey(r.StartKey), convertKey(r.EndKey)
		if l := fieldLen(r.StartKey); l > maxKeyLen {
			maxKeyLen = l
		}
		if l := fieldLen(r.EndKey); l > maxKeyLen {
			maxKeyLen = l
		}
	}
	var maxRegionIDLen int
	for _, r := range regions {
		if l := fieldLen(r.ID); l > maxRegionIDLen {
			maxRegionIDLen = l
		}
	}
	sort.Slice(stores, func(i, j int) bool { return stores[i].Store.ID < stores[j].Store.ID })
	var storeLen []int
	for _, s := range stores {
		storeLen = append(storeLen, fieldLen(s.Store.ID))
	}

	field(maxRegionIDLen, "", "")
	for i := range stores {
		field(storeLen[i], "S"+strconv.FormatUint(stores[i].Store.ID, 10), "")
	}
	field(maxKeyLen, "start", "")
	field(maxKeyLen, "end", "")
	fmt.Println()

	for _, region := range regions {
		field(maxRegionIDLen, "R"+strconv.FormatUint(region.ID, 10), "")
	STORE:
		for i, s := range stores {
			if s.Store.ID == region.Leader.GetStoreId() {
				field(storeLen[i], "▀", "\u001b[31m")
				continue
			}
			for _, p := range region.Peers {
				if p.StoreId == s.Store.ID {
					if p.IsLearner {
						field(storeLen[i], "▀", "\u001b[33m")
					} else {
						field(storeLen[i], "▀", "\u001b[34m")
					}
					continue STORE
				}
			}
			field(storeLen[i], "", "")
		}
		field(maxKeyLen, region.StartKey, "")
		field(maxKeyLen, region.EndKey, "")
		fmt.Println()
	}
}

func convertKey(k string) string {
	b, err := hex.DecodeString(k)
	if err != nil {
		return k
	}
	_, d, err := table.DecodeBytes(b)
	if err != nil {
		return k
	}
	return strings.ToUpper(hex.EncodeToString(d))
}

func fieldLen(f interface{}) int {
	return len(fmt.Sprintf("%v", f)) + 2
}

func field(l int, s string, color string) {
	slen := utf8.RuneCountInString(s)
	if slen > l {
		fmt.Print(s[:l])
		return
	}
	if slen < l {
		fmt.Print(strings.Repeat(" ", l-slen))
	}
	if color != "" {
		fmt.Print(color)
	}
	fmt.Print(s)
	if color != "" {
		fmt.Print("\u001b[0m")
	}
}
