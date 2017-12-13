package blockchain

type Response struct {
	Op string `json:"op"`
	Data interface{} `json:"x"`
}


type Transaction struct {
	LockTime int `json:"lock_time"`
	Version int `json:"ver"`
	Size int `json:"size"`
	Inputs []struct {
		Sequence int64 `json:"sequence"`
		Out struct {
			Spent bool `json:"spent"`
			Index int64 `json:"tx_index"`
			Type int `json:"type"`
			Address string `json:"address"`
			Value int64 `json:"value"`
			Script string `json:"script"`
		} `json:"prev_out"`
	} `json:"inputs"`

	Timestamp int `json:"time"`
	Index int64 `json:"tx_index"`
	VinSize int64 `json:"vin_sz"`
	Hash string `json:"hash"`
	RelayedBy string `json:"relayed_by"`

	Outputs []struct {
		Spent bool `json:"spent"`
		Index int64 `json:"tx_index"`
		Type int `json:"type"`
		Address string `json:"address"`
		Value int64 `json:"value"`
		Script string `json:"script"`
	} `json:"outputs"`
}


type BasicRequest struct {
	Op string `json:"op"`
}