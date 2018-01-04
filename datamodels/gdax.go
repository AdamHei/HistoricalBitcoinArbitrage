package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"github.com/adamhei/historicaldata/models"
	"log"
	"net/http"
	"time"
)

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

func PollGdaxHistorical(interval Interval) ([]PricePoint, *errorhandling.MyError) {
	requestString := buildRequest(interval)
	resp, err := http.Get(requestString)

	if err != nil {
		log.Println("Could not reach ", requestString)
		log.Fatal(err)
	}

	if resp.StatusCode == 200 {
		buckets := make([][]float64, 0)
		err = json.NewDecoder(resp.Body).Decode(&buckets)

		if err != nil {
			return nil, &errorhandling.MyError{Err: err.Error()}
		} else {
			return generalizeGdaxBuckets(buckets), nil
		}
	} else {
		errResp := new(models.GdaxError)
		err = json.NewDecoder(resp.Body).Decode(errResp)

		if err != nil {
			log.Println("Could not decode GDAX error response")
			return nil, &errorhandling.MyError{Err: err.Error()}
		} else {
			return nil, &errorhandling.MyError{Err: errResp.Message}
		}
	}
}

// Convert an array of GdaxBuckets to the more general PricePoints
func generalizeGdaxBuckets(buckets [][]float64) []PricePoint {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		pricePoints[index] = PricePoint{int64(val[0]), val[1]}
	}

	return pricePoints
}

// Given a time interval, buildRequest returns the formatted GET request URL for the GDAX API
// Ex: https://api.gdax.com/products/BTC-USD/candles?start=2017-01-15&end=2017-01-16&granularity=3600
func buildRequest(interval Interval) string {
	req, err := http.NewRequest("GET", historicalEndpoint, nil)
	if err != nil {
		log.Println("Could not build GDAX historical URL")
		log.Fatal(err)
	}

	// Build the GET request
	q := req.URL.Query()

	granularity := getGranularity(interval)
	q.Add("granularity", fmt.Sprintf("%d", granularity))

	now := time.Now()
	startTime := getStartTime(now, interval)
	q.Add("start", startTime.Format("2006-01-02"))
	q.Add("end", now.Format("2006-01-02"))

	req.URL.RawQuery = q.Encode()
	return req.URL.String()
}

// getStartTime returns the time that is interval-before the given time
func getStartTime(now time.Time, interval Interval) time.Time {
	switch interval {
	case TWOYEAR:
		return now.AddDate(-2, 0, 0)
	case YEAR:
		return now.AddDate(-1, 0, 0)
	case SIXMONTH:
		return now.AddDate(0, -6, 0)
	case THREEMONTH:
		return now.AddDate(0, -3, 0)
	case MONTH:
		return now.AddDate(0, -1, 0)
	case WEEK:
		return now.Add(time.Hour * 24 * 7 * -1)
	case DAY:
		return now.Add(time.Hour * 24 * -1)
	default:
		return now.AddDate(-1, 0, 0)
	}
}

// Given a time interval, getGranularity returns the appropriate granularity for the API request
func getGranularity(interval Interval) int {
	switch interval {
	case TWOYEAR:
		return daily
	case YEAR:
		return daily
	case SIXMONTH:
		return daily
	case THREEMONTH:
		return daily
	case MONTH:
		return daily
	case WEEK:
		return fifteenminute
	case DAY:
		return fiveminute
	default:
		return daily
	}
}
