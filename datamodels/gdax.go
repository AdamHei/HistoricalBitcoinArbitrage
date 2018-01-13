package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"github.com/adamhei/historicaldata/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Used when partitioning an interval
type timePeriod struct {
	start, end time.Time
}

var gdaxIntervalToGranularity = map[string]int64{
	TWOYEAR:    dailyBySeconds,
	YEAR:       dailyBySeconds,
	SIXMONTH:   dailyBySeconds,
	THREEMONTH: dailyBySeconds,
	MONTH:      dailyBySeconds,
	WEEK:       fifteenminuteBySeconds,
	DAY:        fiveminuteBySeconds,
}

// GDAX API granularities, with the second being the atomic element
const (
	dailyBySeconds         = 86400
	sixhourBySeconds       = 21600
	hourBySeconds          = 3600
	fifteenminuteBySeconds = 900
	fiveminuteBySeconds    = 300
	minuteBySeconds        = 60
)

const gdaxHistoricalEndpoint = "https://api.gdax.com/products/BTC-USD/candles"

// Given an interval, check its validity and attempt to return all GDAX BTC data within that interval, with a
// pre-determined granularity
func PollGdaxHistorical(interval string) ([]PricePoint, *errorhandling.MyError) {
	interval = strings.ToUpper(interval)
	if gdaxIntervalToGranularity[string(interval)] == 0 {
		return nil, &errorhandling.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: 400}
	}

	buckets, myerror := fetchGdaxBuckets(interval)

	if myerror != nil {
		return nil, myerror
	}
	return generalizeGdaxBuckets(buckets), nil
}

// Convert an array of GdaxBuckets to the more general PricePoints
func generalizeGdaxBuckets(buckets [][]float64) []PricePoint {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		price := strconv.FormatFloat(val[1], 'f', -1, 64)
		pricePoints[index] = PricePoint{int64(val[0]), price}
	}

	return pricePoints
}

// Given a time interval, return a slice of timestamps and BTC prices from GDAX within that interval looking back from today
//
// Some time intervals, such as 2 years and 1 year, require multiple requests to GDAX,
// which is why we treat the intervalPartition as a slice of an arbitrary number of timePeriods/requests to make
func fetchGdaxBuckets(interval string) ([][]float64, *errorhandling.MyError) {
	intervalPartition := getIntervalPartition(interval)
	granularity := gdaxIntervalToGranularity[interval]

	buckets := make([][]float64, 0)
	for _, timePeriod := range intervalPartition {
		requestString, err := buildGdaxRequest(granularity, timePeriod.start, timePeriod.end)

		if err != nil {
			return nil, &errorhandling.MyError{Err: err.Error()}
		}

		response, err := http.Get(requestString)

		if err != nil {
			log.Println("Could not reach ", requestString)
			response.Body.Close()
			return nil, &errorhandling.MyError{Err: "Failed to reach GDAX API", ErrorCode: http.StatusInternalServerError}
		}
		if response.StatusCode == http.StatusOK {
			tempBuckets := make([][]float64, 0)
			err = json.NewDecoder(response.Body).Decode(&tempBuckets)
			response.Body.Close()

			if err != nil {
				log.Println("Could not decode GDAX response")
				return nil, &errorhandling.MyError{Err: err.Error(), ErrorCode: http.StatusInternalServerError}
			}
			buckets = append(buckets, tempBuckets...)
		} else {
			errResp := new(models.GdaxError)
			err = json.NewDecoder(response.Body).Decode(errResp)

			response.Body.Close()
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

// Given a granularity and start and end times, buildGdaxRequest returns the formatted GET request URL for the GDAX API
// Ex: https://api.gdax.com/products/BTC-USD/candles?start=2017-01-15&end=2017-01-16&granularity=3600
func buildGdaxRequest(granularity int64, start time.Time, end time.Time) (string, error) {
	req, err := http.NewRequest("GET", gdaxHistoricalEndpoint, nil)
	if err != nil {
		log.Println("Could not build GDAX historical URL")
		return "", err
	}

	// Build the GET request
	q := req.URL.Query()

	q.Add("granularity", strconv.FormatInt(granularity, 10))

	q.Add("start", start.Format("2006-01-02"))
	q.Add("end", end.Format("2006-01-02"))

	req.URL.RawQuery = q.Encode()
	return req.URL.String(), nil
}

// Given an interval, return a slice partition of that interval into timePeriods in reverse chronological order
// to preserve order when making consecutive requests to GDAX
func getIntervalPartition(interval string) []timePeriod {
	nowRounded := roundTime(time.Now())
	nowRounded = nowRounded.AddDate(0, 0, 1)
	intervalPartition := make([]timePeriod, 0)

	switch interval {
	case TWOYEAR:
		twoYearsAgo := nowRounded.AddDate(-2, 0, 0)
		for timeIndex := nowRounded.AddDate(0, -6, 0); timeIndex.After(twoYearsAgo) || timeIndex.Equal(twoYearsAgo); timeIndex = timeIndex.AddDate(0, -6, 0) {
			intervalPartition = append(intervalPartition, timePeriod{timeIndex, timeIndex.AddDate(0, 6, 0)})
		}
	case YEAR:
		oneYearAgo := nowRounded.AddDate(-1, 0, 0)
		for timeIndex := nowRounded.AddDate(0, -6, 0); timeIndex.After(oneYearAgo) || timeIndex.Equal(oneYearAgo); timeIndex = timeIndex.AddDate(0, -6, 0) {
			intervalPartition = append(intervalPartition, timePeriod{timeIndex, timeIndex.AddDate(0, 6, 0)})
		}
	case SIXMONTH:
		intervalPartition = []timePeriod{{nowRounded.AddDate(0, -6, 0), nowRounded}}
	case THREEMONTH:
		intervalPartition = []timePeriod{{nowRounded.AddDate(0, -3, 0), nowRounded}}
	case MONTH:
		intervalPartition = []timePeriod{{nowRounded.AddDate(0, -1, 0), nowRounded}}
	case WEEK:
		first := nowRounded.AddDate(0, 0, -8)
		second := first.AddDate(0, 0, 3)
		third := second.AddDate(0, 0, 3)
		fourth := third.AddDate(0, 0, 2)
		intervalPartition = []timePeriod{{third, fourth}, {second, third}, {first, second}}
	case DAY:
		intervalPartition = []timePeriod{{nowRounded.AddDate(0, 0, -1), nowRounded}}
	}

	return intervalPartition
}
