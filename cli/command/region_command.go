package command

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

// NewHotRegionCommand
func NewRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "show  region info of the cluster",
	}
	cmd.AddCommand(newRegionPrintCommand())
	return cmd
}

// newRegionPrintCommand
func newRegionPrintCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "print  regions info ",
		Run:   PrintRegionsInfo,
	}
	return cmd
}

func PrintRegionsInfo(cmd *cobra.Command, _ []string) {
	stores, err := GetStoresInfo(cmd)
	if err != nil {
		cmd.Printf("get stores info error :%s \n", err)
	}
	regions, err := GetRegionsInfo(cmd)
	if err != nil {
		cmd.Printf("get regions info error :%s \n", err)
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
			if region.Leader != nil && s.Store.ID == region.Leader.StoreId {
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
	d, ok := decodeBytes(b)
	if !ok {
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

func decodeBytes(b []byte) ([]byte, bool) {
	var buf bytes.Buffer
	for len(b) >= 9 {
		if p := 0xff - b[8]; p >= 0 && p <= 8 {
			buf.Write(b[:8-p])
			b = b[9:]
		} else {
			return nil, false
		}
	}
	return buf.Bytes(), len(b) == 0
}
