package datamodels

import "time"

// Accepted Intervals
const (
	TWOYEAR      = "TWOYEAR"
	YEAR         = "YEAR"
	SIXMONTH     = "SIXMONTH"
	THREEMONTH   = "THREEMONTH"
	MONTH        = "MONTH"
	WEEK         = "WEEK"
	DAY          = "DAY"
	TWELVEHOUR   = "TWELVEHOUR"
	SIXHOUR      = "SIXHOUR"
	HOUR         = "HOUR"
	THIRTYMINUTE = "THIRTYMINUTE"
)

const EMPTYSTRING = ""

const quandlEndpoint = "https://www.quandl.com/api/%s/datasets/%s/%s.json"

// The uniform data structure returned to the client independent of exchange
type PricePoint struct {
	Timestamp int64  `json:"timestamp"`
	Price     string `json:"price"`
}

// Fix to ensure all timestamps returned to the client align on each 5-minute step
func roundTime(t time.Time) time.Time {
	return t.Truncate(time.Minute * 5)
}
