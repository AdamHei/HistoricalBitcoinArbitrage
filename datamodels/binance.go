package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errors"
	"log"
	"net/http"
	"strings"
	"time"
)

// May need in the future if more fields are to be added to PricePoint
// Represents a Binance response bucket

//type binanceResponse struct {
//	OpenTime                 int64
//	Open                     string
//	High                     string
//	Low                      string
//	Close                    string
//	Volumne                  string
//	CloseTime                int64
//	QuoteAssetVolume         string
//	NumTrades                int64
//	TakerBuyBaseAssetVolume  string
//	TakerButQuoteAssetVolume string
//	Ignore                   string
//}

type binanceError struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

// The default granularity for each interval
var binanceIntervals = map[string]string{
	TWOYEAR:    "1d",
	YEAR:       "1d",
	SIXMONTH:   "1d",
	THREEMONTH: "1d",
	MONTH:      "1d",
	WEEK:       "6h",
	DAY:        "15m",
}

const BTCUSDT = "BTCUSDT"

const binanceApiVersion = "v1"
const binanceHistoricalEndpoint = "https://api.binance.com/api/%s/klines"

var binanceEndpoint = fmt.Sprintf(binanceHistoricalEndpoint, binanceApiVersion)

// Given an interval, check its validity and return all open prices within that interval and any relevant errors
func PollBinanceHistorical(interval string) ([]PricePoint, *errors.MyError) {
	interval = strings.ToUpper(interval)
	if binanceIntervals[string(interval)] == interval {
		return nil, &errors.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: 400}
	}

	buckets, myerror := fetchBinanceBuckets(interval)

	if myerror != nil {
		return nil, myerror
	}

	return unmarshalBinanceBuckets(buckets), nil
}

// Binance gives us JSON arrays of mixed strings and integers, which makes parsing unnecessarily difficult
func unmarshalBinanceBuckets(buckets [][]interface{}) []PricePoint {
	numBuckets := len(buckets)
	pricePoints := make([]PricePoint, numBuckets)

	for index, val := range buckets {
		timestamp := val[0].(float64)
		// Convert millis -> seconds
		timestamp = timestamp / 1000

		open := val[1].(string)

		// Binance gives us data in ascending order, so we must reverse!
		pricePoints[numBuckets-1-index] = PricePoint{Timestamp: int64(timestamp), Price: open}
	}

	return pricePoints
}

// Attempt to build the Binance request, query for the data, return the raw data if sucessful and any errors else
func fetchBinanceBuckets(interval string) ([][]interface{}, *errors.MyError) {
	requestString, err := buildBinanceRequest(interval)
	if err != nil {
		return nil, &errors.MyError{Err: err.Error()}
	}

	response, err := http.Get(requestString)
	defer response.Body.Close()

	log.Println(fmt.Sprintf("Querying %s", requestString))

	if err != nil {
		log.Println(fmt.Sprintf("Could not reach %s", requestString))
		return nil, &errors.MyError{Err: err.Error()}
	}

	buckets := make([][]interface{}, 0)
	if response.StatusCode == http.StatusOK {
		tempBuckets := make([][]interface{}, 0)
		err = json.NewDecoder(response.Body).Decode(&tempBuckets)

		if err != nil {
			log.Println("Could not decode Binance response")
			return nil, &errors.MyError{Err: err.Error(), ErrorCode: http.StatusInternalServerError}
		}

		buckets = append(buckets, tempBuckets...)
	} else {
		binanceErr := new(binanceError)
		err = json.NewDecoder(response.Body).Decode(binanceErr)

		if err != nil {
			log.Println("Could not decode Binance error response")
			return nil, &errors.MyError{Err: err.Error(), ErrorCode: http.StatusInternalServerError}
		}
		return nil, &errors.MyError{Err: binanceErr.Msg, ErrorCode: http.StatusInternalServerError}
	}

	return buckets, nil
}

// Given an interval, construct the proper GET request with all properly formatted params
func buildBinanceRequest(interval string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, binanceEndpoint, nil)
	if err != nil {
		log.Println("Could not build Binance request")
		return EMPTYSTRING, nil
	}

	query := request.URL.Query()

	query.Add("symbol", BTCUSDT)
	query.Add("interval", binanceIntervals[interval])
	query.Add("startTime", fmt.Sprintf("%d", getBinanceStartDate(interval)))

	request.URL.RawQuery = query.Encode()
	return request.URL.String(), nil
}

// Based on the interval, return the start time in milliseconds
func getBinanceStartDate(interval string) int64 {
	startTime := time.Now()

	switch interval {
	case TWOYEAR:
		startTime = startTime.AddDate(-2, 0, 0)
	case YEAR:
		startTime = startTime.AddDate(-1, 0, 0)
	case SIXMONTH:
		startTime = startTime.AddDate(0, -6, 0)
	case THREEMONTH:
		startTime = startTime.AddDate(0, -3, 0)
	case MONTH:
		startTime = startTime.AddDate(0, -1, 0)
	case WEEK:
		startTime = startTime.AddDate(0, 0, -7)
	case DAY:
		startTime = startTime.AddDate(0, 0, -1)
	}

	// Convert seconds to millis
	return startTime.Unix() * 1000
}
