package coindesk

import "time"

type CurrentPrice struct {
	Time       Time   `json:"time"`
	Disclaimer string `json:"disclaimer"`
	Bpi        BpiMap `json:"bpi"`
}

type Time struct {
	Updated    string    `json:"updated"`
	UpdatedISO time.Time `json:"updatedISO"`
	Updateduk  string    `json:"updateduk"`
}

type BpiMap map[string]Bpi

type Bpi struct {
	Code        string  `json:"code"`
	Rate        string  `json:"rate"`
	Description string  `json:"description"`
	RateFloat   float64 `json:"rate_float"`
}
