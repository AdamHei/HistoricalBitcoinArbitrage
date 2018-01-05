package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"github.com/adamhei/historicaldata/models"
	"log"
	"net/http"
	"strings"
	"time"
)

// Used when partitioning an interval
type timePeriod struct {
	start, end time.Time
}

var intervalToGranularity = map[string]int{
	TWOYEAR:    daily,
	YEAR:       daily,
	SIXMONTH:   daily,
	THREEMONTH: daily,
	MONTH:      daily,
	WEEK:       fifteenminute,
	DAY:        fiveminute,
}

// GDAX API granularities
const (
	daily         = 86400
	sixhour       = 21600
	hour          = 3600
	fifteenminute = 900
	fiveminute    = 300
	minute        = 60
)

const historicalEndpoint = "https://api.gdax.com/products/BTC-USD/candles"

// Given an interval, check its validity and attempt to return all GDAX BTC data within that interval, with a
// pre-determined granularity
func PollGdaxHistorical(interval string) ([]PricePoint, *errorhandling.MyError) {
	interval = strings.ToUpper(interval)
	if intervalToGranularity[string(interval)] == 0 {
		return nil, &errorhandling.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: 400}
	}

	buckets, myerror := fetchBuckets(interval)

	if myerror != nil {
		return nil, myerror
	}
	return generalizeGdaxBuckets(buckets), nil
}

// Convert an array of GdaxBuckets to the more general PricePoints
func generalizeGdaxBuckets(buckets [][]float64) []PricePoint {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		pricePoints[index] = PricePoint{int64(val[0]), val[1]}
	}

	return pricePoints
}

// Given a time interval, return a slice of timestamps and BTC prices from GDAX within that interval looking back from today
//
// Some time intervals, such as 2 years and 1 year, require multiple requests to GDAX,
// which is why we treat the intervalPartition as a slice of an arbitrary number of timePeriods/requests to make
func fetchBuckets(interval string) ([][]float64, *errorhandling.MyError) {
	intervalPartition := getIntervalPartition(interval)
	granularity := intervalToGranularity[interval]

	buckets := make([][]float64, 0)
	for _, timePeriod := range intervalPartition {
		requestString, err := buildRequest(granularity, timePeriod.start, timePeriod.end)

		if err != nil {
			return nil, &errorhandling.MyError{Err: err.Error()}
		}

		response, err := http.Get(requestString)

		if err != nil {
			log.Println("Could not reach ", requestString)
			return nil, &errorhandling.MyError{Err: "Failed to reach GDAX API", ErrorCode: http.StatusInternalServerError}
		}
		if response.StatusCode == http.StatusOK {
			tempBuckets := make([][]float64, 0)
			err = json.NewDecoder(response.Body).Decode(&tempBuckets)

			if err != nil {
				log.Println("Could not decode GDAX response")
				return nil, &errorhandling.MyError{Err: err.Error(), ErrorCode: http.StatusInternalServerError}
			}
			buckets = append(buckets, tempBuckets...)
		} else {
			errResp := new(models.GdaxError)
			err = json.NewDecoder(response.Body).Decode(errResp)

			if err != nil {
				log.Println("Could not decode GDAX error response with code ", response.StatusCode)
				return nil, &errorhandling.MyError{Err: err.Error()}
			} else {
				return nil, &errorhandling.MyError{Err: errResp.Message}
			}
		}
	}

	return buckets, nil
}

// Given a granularity and start and end times, buildRequest returns the formatted GET request URL for the GDAX API
// Ex: https://api.gdax.com/products/BTC-USD/candles?start=2017-01-15&end=2017-01-16&granularity=3600
func buildRequest(granularity int, start time.Time, end time.Time) (string, error) {
	req, err := http.NewRequest("GET", historicalEndpoint, nil)
	if err != nil {
		log.Println("Could not build GDAX historical URL")
		return "", err
	}

	// Build the GET request
	q := req.URL.Query()

	q.Add("granularity", fmt.Sprintf("%d", granularity))

	q.Add("start", start.Format("2006-01-02"))
	q.Add("end", end.Format("2006-01-02"))

	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}

// Given an interval, return a slice partition of that interval into timePeriods in reverse chronological order
// to preserve order when making consecutive requests to GDAX
func getIntervalPartition(interval string) []timePeriod {
	now := time.Now()
	now = now.AddDate(0, 0, 1)
	intervalPartition := make([]timePeriod, 0)

	switch interval {
	case TWOYEAR:
		twoYearsAgo := now.AddDate(-2, 0, 0)
		for timeIndex := now.AddDate(0, -6, 0); timeIndex.After(twoYearsAgo) || timeIndex.Equal(twoYearsAgo); timeIndex = timeIndex.AddDate(0, -6, 0) {
			intervalPartition = append(intervalPartition, timePeriod{timeIndex, timeIndex.AddDate(0, 6, 0)})
		}
	case YEAR:
		oneYearAgo := now.AddDate(-1, 0, 0)
		for timeIndex := now.AddDate(0, -6, 0); timeIndex.After(oneYearAgo) || timeIndex.Equal(oneYearAgo); timeIndex = timeIndex.AddDate(0, -6, 0) {
			intervalPartition = append(intervalPartition, timePeriod{timeIndex, timeIndex.AddDate(0, 6, 0)})
		}
	case SIXMONTH:
		intervalPartition = []timePeriod{{now.AddDate(0, -6, 0), now}}
	case THREEMONTH:
		intervalPartition = []timePeriod{{now.AddDate(0, -3, 0), now}}
	case MONTH:
		intervalPartition = []timePeriod{{now.AddDate(0, -1, 0), now}}
	case WEEK:
		first := now.AddDate(0, 0, -8)
		second := first.AddDate(0, 0, 3)
		third := second.AddDate(0, 0, 3)
		fourth := third.AddDate(0, 0, 2)
		intervalPartition = []timePeriod{{third, fourth}, {second, third}, {first, second}}
	case DAY:
		intervalPartition = []timePeriod{{now.AddDate(0, 0, -1), now}}
	}

	return intervalPartition
}
