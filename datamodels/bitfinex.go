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

const bitfinex = "BITFINEX"
const bitfinexTicker = "BTCUSD"
const qBitfinexTemplate = "https://www.quandl.com/api/%s/datasets/%s/%s.json"

var qBitfinexEndpoint = fmt.Sprintf(qBitfinexTemplate, quandlApiV3, bitfinex, bitfinexTicker)

// Representation of a single Quandl data bucket, e.g.
type qBitfinexBucket struct {
	Date                                   string
	High, Low, Mid, Last, Bid, Ask, Volume float64
}

// Given an interval, check its validity and return all Bitfinex data within that interval, as PricePoints
// Currently, we only support Bitfinex data through Quandl, whose finest granularity is one day
//
// TODO: Add direct Bitfinex API to support intervals shorter than 1 month
func PollBitfinexHistorical(interval string) ([]PricePoint, *errors.MyError) {
	interval = strings.ToUpper(interval)
	if !quandlIntervals[interval] {
		return nil, &errors.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: http.StatusBadRequest}
	}

	requestString, err := buildQBitfinexRequest(interval)
	if err != nil {
		return nil, &errors.MyError{Err: err.Error()}
	}

	quandlResponse, myErr := fetchQuandlResponse(requestString)
	if err != nil {
		return nil, myErr
	}

	return parseQBitfinexBuckets(quandlResponse.DataSetResponse.Data)
}

// Given the raw 2D Quandl data, convert it to an array of PricePoints
func parseQBitfinexBuckets(buckets [][]json.RawMessage) ([]PricePoint, *errors.MyError) {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		bucket, err := unmarshalQBitfinexBucket(val)
		if err != nil {
			return nil, &errors.MyError{Err: err.Error()}
		}

		timestamp, err := time.Parse(DATELAYOUTSTRING, bucket.Date)
		if err != nil {
			log.Println("Could not parse time from Bitfinex bucket")
			return nil, &errors.MyError{Err: "Failure to parse Quandl response", ErrorCode: http.StatusInternalServerError}
		}

		price := strconv.FormatFloat(bucket.Mid, 'f', -1, 64)

		pricePoints[index] = PricePoint{Timestamp: timestamp.Unix(), Price: price}
	}

	return pricePoints, nil
}

// Quandl decided to return an array of both strings and floats, forcing us to parse it by hand
// For now, just timestamp and mid price
//
// TODO: Add more fields as necessary
func unmarshalQBitfinexBucket(jsonBucket []json.RawMessage) (*qBitfinexBucket, error) {
	bucket := new(qBitfinexBucket)

	// Parse date
	err := json.Unmarshal(jsonBucket[0], &bucket.Date)
	if err != nil {
		return nil, err
	}

	// Parse mid-price
	err = json.Unmarshal(jsonBucket[3], &bucket.Mid)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}



// Given an interval, add the custom GET parameters to the Quandl request
func buildQBitfinexRequest(interval string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, qBitfinexEndpoint, nil)
	if err != nil {
		log.Println("Could not build Quandl-Bitfinex URL")
		return EMPTYSTRING, err
	}

	query := request.URL.Query()

	query.Add("api_key", quandlApiKey)
	query.Add("start_date", getQuandlStartDate(interval))

	request.URL.RawQuery = query.Encode()
	return request.URL.String(), nil
}
