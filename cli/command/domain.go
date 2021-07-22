package command

type Peer struct {
	Id        uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	StoreId   uint64 `protobuf:"varint,2,opt,name=store_id,json=storeId,proto3" json:"store_id,omitempty"`
	IsLearner bool   `protobuf:"varint,3,opt,name=is_learner,json=isLearner,proto3" json:"is_learner,omitempty"`
}

type RegionInfo struct {
	ID       uint64  `json:"id"`
	StartKey string  `json:"start_key"`
	EndKey   string  `json:"end_key"`
	Peers    []*Peer `json:"peers,omitempty"`
	Leader   *Peer   `json:"leader,omitempty"`
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
