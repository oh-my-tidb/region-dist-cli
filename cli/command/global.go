package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func getEndpoints(cmd *cobra.Command) []string {
	addrs, err := cmd.Flags().GetString("pd")
	if err != nil {
		cmd.Println("get pd address failed, should set flag with '-u'")
		os.Exit(1)
	}
	eps := strings.Split(addrs, ",")
	for i, ep := range eps {
		if j := strings.Index(ep, "//"); j == -1 {
			eps[i] = "//" + ep
		}
	}
	return eps
}

// GetStoresInfo
func GetStoresInfo(cmd *cobra.Command) (*StoresInfo, error) {
	addrs := getEndpoints(cmd)
	res, err := http.Get(addrs[0] + "/pd/api/v1/stores")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var stores StoresInfo
	err = json.NewDecoder(res.Body).Decode(&stores)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &stores, nil
}

func GetRegionsInfo(cmd *cobra.Command) (*RegionsInfo, error) {
	addrs := getEndpoints(cmd)
	res, err := http.Get(addrs[0] + "/pd/api/v1/regions")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var regions RegionsInfo
	err = json.NewDecoder(res.Body).Decode(&regions)
	if err != nil {
		return nil, err
	}
	return &regions, nil
}
