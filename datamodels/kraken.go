package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Top level Kraken response body
type KrakenResponse struct {
	Error  []string        `json:"error"`
	Result KrakenResultMap `json:"result"`
}

// Mid-level Kraken response body containing actual price data
type KrakenResultMap struct {
	Buckets [][]json.RawMessage `json:"XXBTZUSD"`
	Last    int64               `json:"last"`
}

// Represents an individual array of instantaneous price data
type KrakenBucket struct {
	Timestamp                            int64
	Open, High, Low, Close, Vwap, Volume string
	Count                                int64
}

//Open-High-Low-Close intervals with minute being the atomic element
const (
	minute               = 1
	fiveMinutes          = 5
	fifteenMinutes       = 15
	thirtyMinutes        = 30
	hourByMinutes        = 60
	fourHourByMinutes    = 240
	oneDayByMinutes      = 1440
	oneWeekByMinutes     = 10080
	fifteenDaysByMinutes = 21600
)

// Represents which intervals are supported and their corresponding granularity
var krakenIntervalToGranularity = map[string]int64{
	TWOYEAR:    oneDayByMinutes,
	YEAR:       oneDayByMinutes,
	SIXMONTH:   oneDayByMinutes,
	THREEMONTH: oneDayByMinutes,
	MONTH:      oneDayByMinutes,
	WEEK:       fifteenMinutes,
	DAY:        fiveMinutes,
}

const krakenApiVersion = "0"
const krakenEndpoint = "https://api.kraken.com/%s/public/OHLC"
var krakenHistoricalEndpoint = fmt.Sprintf(krakenEndpoint, krakenApiVersion)

// Given an interval, check its validity and return all Kraken BTC data within that interval, by a pre-determined granularity
func PollKrakenHistorical(interval string) ([]PricePoint, *errors.MyError) {
	interval = strings.ToUpper(interval)
	if krakenIntervalToGranularity[string(interval)] == 0 {
		return nil, &errors.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: http.StatusBadRequest}
	}

	resultMap, err := fetchKrakenResponse(interval)
	if err != nil {
		return nil, err
	}

	return parseKrakenBuckets(resultMap.Buckets)
}

// Given the Kraken 2D-price data array, convert it to an array of the universal PricePoint data structure
func parseKrakenBuckets(buckets [][]json.RawMessage) ([]PricePoint, *errors.MyError) {
	n := len(buckets)
	pricePoints := make([]PricePoint, n)

	for index, val := range buckets {
		bucket := new(KrakenBucket)

		err := unmarshalKrakenBucket(val, bucket)

		if err != nil {
			return nil, &errors.MyError{Err: err.Error()}
		}

		// Return the pricepoints in descending order (newest to oldest)
		pricePoints[n-1-index] = PricePoint{bucket.Timestamp, bucket.Open}
	}
	return pricePoints, nil
}

// Since Kraken decided to return an array of strings and integers, we need to parse it by hand
// For now, just parsing timestamp and open price
//
// TODO: Add more fields as necessary
func unmarshalKrakenBucket(jsonBucket []json.RawMessage, bucket *KrakenBucket) error {
	err := json.Unmarshal(jsonBucket[0], &bucket.Timestamp)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBucket[1], &bucket.Open)
	if err != nil {
		return err
	}

	return nil
}

// Given an interval:
// 1. Construct the GET request
// 2. Fetch the historical data from Kraken
// 3. Return KrakenResultMap if successful, error else
func fetchKrakenResponse(interval string) (*KrakenResultMap, *errors.MyError) {
	requestString, err := buildKrakenRequest(interval)

	if err != nil {
		log.Println("Could build Kraken request string")
		return nil, &errors.MyError{Err: err.Error()}
	}

	response, err := http.Get(requestString)
	defer response.Body.Close()

	log.Println(fmt.Sprintf("Querying %s", requestString))

	if err != nil {
		log.Println("Could not reach ", requestString)
		return nil, &errors.MyError{Err: err.Error()}
	}

	if response.StatusCode == http.StatusOK {
		krakenResponse := new(KrakenResponse)
		err = json.NewDecoder(response.Body).Decode(krakenResponse)

		if err != nil {
			log.Println("Could not decode Kraken Response")
			return nil, &errors.MyError{Err: err.Error()}
		}
		if len(krakenResponse.Error) > 0 {
			log.Println(krakenResponse.Error[0])
			return nil, &errors.MyError{Err: krakenResponse.Error[0]}
		}
		return &krakenResponse.Result, nil
	} else {
		resp := new(interface{})
		json.NewDecoder(response.Body).Decode(&resp)
		log.Println(resp)
		log.Println(fmt.Sprintf("Either the Kraken API is down or the request was incorrect with response code %d", response.StatusCode))
		return nil, &errors.MyError{Err: "Kraken API error", ErrorCode: http.StatusInternalServerError}
	}
}

// From an interval, add the custom GET parameters to the Kraken request
func buildKrakenRequest(interval string) (string, error) {
	request, err := http.NewRequest("GET", krakenHistoricalEndpoint, nil)
	if err != nil {
		log.Println("Could not build Kraken historical URL")
		return EMPTYSTRING, err
	}

	query := request.URL.Query()

	query.Add("pair", "XXBTZUSD")
	query.Add("interval", strconv.FormatInt(krakenIntervalToGranularity[interval], 10))

	since := getRoundedStartTime(interval)

	query.Add("since", strconv.FormatInt(since.Unix(), 10))

	request.URL.RawQuery = query.Encode()
	return request.URL.String(), nil
}

// We round the time to the nearest 5-min step to synchronize consecutive requests
func getRoundedStartTime(interval string) time.Time {
	startTime := time.Now()
	startTime = roundTime(startTime)

	switch interval {
	case TWOYEAR:
		return startTime.AddDate(-2, 0, 0)
	case YEAR:
		return startTime.AddDate(-1, 0, 0)
	case SIXMONTH:
		return startTime.AddDate(0, -6, 0)
	case THREEMONTH:
		return startTime.AddDate(0, -3, 0)
	case MONTH:
		return startTime.AddDate(0, -3, 0)
	case WEEK:
		return startTime.AddDate(0, 0, -8)
	case DAY:
		return startTime.AddDate(0, 0, -1)
	default:
		return startTime.AddDate(-1, 0, 0)
	}
}
