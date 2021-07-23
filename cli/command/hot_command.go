package command

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/spf13/cobra"
)

// NewHotRegionCommand
func NewHotRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hot",
		Short: "show hot region info of the cluster",
	}
	cmd.AddCommand(newHotExportCommand())
	return cmd
}

func newHotExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "export regions info ",
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
		topReadPath: "/pd/api/v1/" + "/hotspot/regions/read",
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

			// set read metrics
			for k, v := range []float64{region.ByteRate, region.KeyRate, region.QueryRate} {
				a = string('B'+count+k+1) + strconv.Itoa(regionCount)
				f.SetCellFloat("hot region", a, v, 2, 32)
			}

			regionInfo := h.regionDic[region.RegionID]
			// set read metrics
			a = string('B'+count+7) + strconv.Itoa(regionCount)
			f.SetCellStr("hot region", a, regionInfo.StartKey)
			a = string('B'+count+8) + strconv.Itoa(regionCount)
			f.SetCellStr("hot region", a, regionInfo.EndKey)

			for _, peer := range regionInfo.Peers {
				index := h.storeDic[peer.StoreId]
				a = string('B'+int8(index)) + strconv.Itoa(regionCount)
				f.SetCellInt("hot region", a, 1)
			}
		}
	}
	f.SetActiveSheet(sheet)
	if err := f.SaveAs("hot.xlsx"); err != nil {
		return err
	}
	return nil
}
