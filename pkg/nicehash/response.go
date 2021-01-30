package nicehash

type GetRigs2Response struct {
	MiningRigs              []MiningRig `json:"miningRigs"`
	ExternalAddress         bool        `json:"externalAddress"`
	TotalProfitabilityLocal float64     `json:"totalProfitabilityLocal"`
	Pagination              Pagenation  `json:"pagination"`
}

type Pagenation struct {
	Size           int `json:"size"`
	Page           int `json:"page"`
	TotalPageCount int `json:"totalPageCount"`
}

type Algorithm struct {
	EnumName    string `json:"enumName"`
	Description string `json:"description"`
}

type MiningRig struct {
	RigID              string `json:"rigId"`
	Type               string `json:"type"`
	Name               string `json:"name"`
	StatusTime         int64  `json:"statusTime"`
	MinerStatus        string `json:"minerStatus"`
	UnpaidAmount       string `json:"unpaidAmount"`
	Stats              []Stat
	Profitability      float64 `json:"profitability"`
	LocalProfitability float64 `json:"localProfitability"`
}

type Stat struct {
	StatsTime                int64     `json:"statsTime"`
	Market                   string    `json:"market"`
	Algorithm                Algorithm `json:"algorithm"`
	UnpaidAmount             string    `json:"unpaidAmount"`
	Difficulty               float64   `json:"difficulty"`
	ProxyID                  int       `json:"proxyId"`
	TimeConnected            int64     `json:"timeConnected"`
	Xnsub                    bool      `json:"xnsub"`
	SpeedAccepted            float64   `json:"speedAccepted"`
	SpeedRejectedR1Target    float64   `json:"speedRejectedR1Target"`
	SpeedRejectedR2Stale     float64   `json:"speedRejectedR2Stale"`
	SpeedRejectedR3Duplicate float64   `json:"speedRejectedR3Duplicate"`
	SpeedRejectedR4NTime     float64   `json:"speedRejectedR4NTime"`
	SpeedRejectedR5Other     float64   `json:"speedRejectedR5Other"`
	SpeedRejectedTotal       float64   `json:"speedRejectedTotal"`
	Profitability            float64   `json:"profitability"`
}
