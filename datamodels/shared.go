// Package datamodels interfaces either directly with each exchange's API or with the local database
//
// Each datamodel must have a function which returns an array of PricePoints and an optional error
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
const DATELAYOUTSTRING = "2006-01-02"

// The uniform data structure returned to the client independent of exchange
// Represents a price at a specific point in time
type PricePoint struct {
	Timestamp int64  `json:"timestamp"`
	Price     string `json:"price"`
}

// Fix to ensure all timestamps returned to the client align on each 5-minute step
func roundTime(t time.Time) time.Time {
	return t.Truncate(time.Minute * 5)
}
