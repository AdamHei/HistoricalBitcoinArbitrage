package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type KrakenResponse struct {
	Error  []string        `json:"error"`
	Result KrakenResultMap `json:"result"`
}

type KrakenResultMap struct {
	Buckets [][]json.RawMessage `json:"XXBTZUSD"`
	Last    int64               `json:"last"`
}

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

var krakenIntervalToGranularity = map[string]int64{
	TWOYEAR:    oneDayByMinutes,
	YEAR:       oneDayByMinutes,
	SIXMONTH:   oneDayByMinutes,
	THREEMONTH: oneDayByMinutes,
	MONTH:      oneDayByMinutes,
	WEEK:       fifteenMinutes,
	DAY:        fiveMinutes,
}

const krakenHistoricalEndpoint = "https://api.kraken.com/0/public/OHLC"

// Given an interval, check its validity and return all Kraken BTC data within that interval, by a pre-determined granularity
func PollKrakenHistorical(interval string) ([]PricePoint, *errorhandling.MyError) {
	interval = strings.ToUpper(interval)
	if krakenIntervalToGranularity[string(interval)] == 0 {
		return nil, &errorhandling.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: 400}
	}

	resultMap, err := fetchKrakenResponse(interval)

	if err != nil {
		return nil, err
	}

	return parseKrakenBuckets(resultMap.Buckets)
}

// Given the Kraken 2D-price data array, convert it to an array of the universal PricePoint data structure
func parseKrakenBuckets(buckets [][]json.RawMessage) ([]PricePoint, *errorhandling.MyError) {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		bucket := new(KrakenBucket)

		err := unmarshalBucket(val, bucket)

		if err != nil {
			return nil, &errorhandling.MyError{Err: err.Error()}
		}

		pricePoints[index] = PricePoint{bucket.Timestamp, bucket.Open}
	}
	return pricePoints, nil
}

// Since Kraken decided to return an array of strings and integers, we need to parse it by hand
// For now, just parsing timestamp and open price
func unmarshalBucket(jsonBucket []json.RawMessage, bucket *KrakenBucket) error {
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
func fetchKrakenResponse(interval string) (*KrakenResultMap, *errorhandling.MyError) {
	requestString, err := buildKrakenRequest(interval)

	if err != nil {
		log.Println("Could build Kraken request string")
		return nil, &errorhandling.MyError{Err: err.Error()}
	}

	response, err := http.Get(requestString)
	log.Println(fmt.Sprintf("Queryed %s", requestString))

	if err != nil {
		log.Println("Could not reach ", requestString)
		return nil, &errorhandling.MyError{Err: err.Error()}
	}

	if response.StatusCode == http.StatusOK {
		krakenResponse := new(KrakenResponse)
		err = json.NewDecoder(response.Body).Decode(krakenResponse)

		if err != nil {
			log.Println("Could not decode Kraken Response")
			return nil, &errorhandling.MyError{Err: err.Error()}
		}
		if len(krakenResponse.Error) > 0 {
			log.Println(krakenResponse.Error[0])
			return nil, &errorhandling.MyError{Err: krakenResponse.Error[0]}
		}
		return &krakenResponse.Result, nil
	} else {
		resp := new(interface{})
		json.NewDecoder(response.Body).Decode(&resp)
		log.Println(resp)
		log.Println(fmt.Sprintf("Either the Kraken API is down or the request was incorrect with response code %d", response.StatusCode))
		return nil, &errorhandling.MyError{Err: "Kraken API error", ErrorCode: 500}
	}
}

// From an interval, add the custom GET parameters to the Kraken request
func buildKrakenRequest(interval string) (string, error) {
	request, err := http.NewRequest("GET", krakenHistoricalEndpoint, nil)
	if err != nil {
		log.Println("Could not build Kraken historical URL")
		return "", err
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
