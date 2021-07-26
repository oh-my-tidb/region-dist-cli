package command

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/spf13/cobra"
)

type hotType string

const (
	// Initiated by admin.
	Read hotType = "/hotspot/regions/read"
	// Initiated by merge checker or merge scheduler. Note that it may not include region merge.
	// the order describe the operator's producer and is very helpful to decouple scheduler or checker limit
	Write hotType = "/hotspot/regions/write"
)

var defaultHotType = Read

// NewHotRegionCommand
func NewHotRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hot",
		Short: "export hot region info of the cluster",
	}
	cmd.AddCommand(
		newReadHotExportCommand(),
		newWriteHotExportCommand())
	return cmd
}

func newReadHotExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "export hot read regions info ",
		Run:   ShowRegionDistributionFnc,
	}
	return cmd
}
func newWriteHotExportCommand() *cobra.Command {
	defaultHotType = Write
	cmd := &cobra.Command{
		Use:   "write",
		Short: "export hot write regions info ",
		Run:   ShowRegionDistributionFnc,
	}
	return cmd
}

func ShowRegionDistributionFnc(cmd *cobra.Command, args []string) {
	stores, err := GetStoresInfo(cmd)
	if err != nil {
		cmd.Printf("get stores info error :%s \n", err)
	}
	regions, err := GetRegionsInfo(cmd)
	if err != nil {
		cmd.Printf("get regions info error :%s \n", err)
	}
	pd := getEndpoints(cmd)
	export := NewHotRegionExport(pd[0], stores, regions)
	err = export.prepare()
	if err != nil {
		cmd.Printf("get regions info error :%s \n", err)
	}
	err = export.export()
	if err != nil {
		cmd.Printf("get regions info error :%s \n", err)
	}
}

type StoreInfos struct {
	StoreHotPeersStat *StoreHotPeersInfos
	topReadPath       string
	pd                string
	storeDic          map[uint64]int
	regionDic         map[uint64]*RegionInfo
}
type StoreHotPeersInfos struct {
	AsPeer   map[string]*HotPeersStat `json:"as_peer"`
	AsLeader map[string]*HotPeersStat `json:"as_leader"`
}

// HotPeersStat records all hot regions statistics
type HotPeersStat struct {
	TotalLoads     []float64         `json:"-"`
	TotalBytesRate float64           `json:"total_flow_bytes"`
	TotalKeysRate  float64           `json:"total_flow_keys"`
	TotalQueryRate float64           `json:"total_flow_query"`
	Count          int               `json:"regions_count"`
	Stats          []HotPeerStatShow `json:"statistics"`
}

// HotPeerStatShow records the hot region statistics for output
type HotPeerStatShow struct {
	StoreID        uint64    `json:"store_id"`
	RegionID       uint64    `json:"region_id"`
	HotDegree      int       `json:"hot_degree"`
	ByteRate       float64   `json:"flow_bytes"`
	KeyRate        float64   `json:"flow_keys"`
	QueryRate      float64   `json:"flow_query"`
	AntiCount      int       `json:"anti_count"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

func NewHotRegionExport(pd string, stores *StoresInfo, regions *RegionsInfo) *StoreInfos {
	storeDic := mapStore(stores)
	regionDic := mapRegion(regions)
	return &StoreInfos{
		pd:          pd,
		topReadPath: "/pd/api/v1/" + string(defaultHotType),
		storeDic:    storeDic,
		regionDic:   regionDic,
	}
}

func (h *StoreInfos) prepare() error {
	res, err := http.Get(h.pd + h.topReadPath)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var s StoreHotPeersInfos
	err = json.NewDecoder(res.Body).Decode(&s)
	if err != nil {
		return err
	}
	h.StoreHotPeersStat = &s
	return nil
}

func mapStore(stores *StoresInfo) map[uint64]int {
	dic := make(map[uint64]int, len(stores.Stores))
	for i, v := range stores.Stores {
		dic[v.Store.ID] = i
	}
	return dic
}

// mapRegion returns the relationship between region leader and stores
func mapRegion(regions *RegionsInfo) map[uint64]*RegionInfo {
	rst := make(map[uint64]*RegionInfo, len(regions.Regions))
	for _, v := range regions.Regions {
		rst[v.ID] = v
	}
	return rst
}

func (h *StoreInfos) export() error {
	f := excelize.NewFile()
	sheet := f.NewSheet("hot region")
	f.SetCellValue("hot region", "A2", "test")
	count := len(h.storeDic)

	a := string('B'+int8(count)) + strconv.Itoa(1)
	f.SetCellStr("hot region", a, "leader")
	for k, v := range []string{"read_bytes", "read_keys", "read_qps", "write_bytes", "write_bytes", "write_qps",
		"start_key", "end_key", "table", "is_index"} {
		a = string('B'+int8(count+k+1)) + strconv.Itoa(1)
		f.SetCellStr("hot region", a, v)
	}

	regionCount := 1
	// record data
	for id, store := range h.StoreHotPeersStat.AsLeader {
		storeID, _ := strconv.Atoi(id)
		a = string('B'+int8(h.storeDic[uint64(storeID)])) + strconv.Itoa(1)
		f.SetCellInt("hot region", a, storeID)
		for _, region := range store.Stats {
			regionCount++
			// set regionID
			a = string('A') + strconv.Itoa(regionCount)
			f.SetCellInt("hot region", a, int(region.RegionID))

			// set region leader
			a = string('B'+count) + strconv.Itoa(regionCount)
			f.SetCellInt("hot region", a, h.storeDic[uint64(storeID)])

