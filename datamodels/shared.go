package datamodels

const (
	TWOYEAR      Interval = "TWOYEAR"
	YEAR         Interval = "YEAR"
	SIXMONTH     Interval = "SIXMONTH"
	THREEMONTH   Interval = "THREEMONTH"
	MONTH        Interval = "MONTH"
	WEEK         Interval = "WEEK"
	DAY          Interval = "DAY"
	TWELVEHOUR   Interval = "TWELVEHOUR"
	SIXHOUR      Interval = "SIXHOUR"
	HOUR         Interval = "HOUR"
	THIRTYMINUTE Interval = "THIRTYMINUTE"
)

type PricePoint struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}
