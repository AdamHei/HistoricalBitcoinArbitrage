package datamodels

// Intervals
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

type PricePoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}
