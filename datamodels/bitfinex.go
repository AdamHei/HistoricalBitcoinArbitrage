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

type QuandlResponse struct {
	DataSetResponse QuandlDataSetResponse `json:"dataset"`
}

type QuandlDataSetResponse struct {
	Id                  int64               `json:"id"`
	DatasetCode         string              `json:"dataset_code"`
	DatabaseCode        string              `json:"database_code"`
	Name                string              `json:"name"`
	Description         string              `json:"description"`
	RefreshedAt         string              `json:"refreshed_at"`
	NewestAvailableDate string              `json:"newest_available_date"`
	OldestAvailableDate string              `json:"oldest_available_date"`
	ColumnNames         []string            `json:"column_names"`
	Frequency           string              `json:"frequency"`
	Type                string              `json:"type"`
	Premium             bool                `json:"premium"`
	Limit               json.RawMessage     `json:"limit"`
	Transform           json.RawMessage     `json:"transform"`
	ColumnIndex         json.RawMessage     `json:"column_index"`
	StartDate           string              `json:"start_date"`
	EndDate             string              `json:"end_date"`
	Data                [][]json.RawMessage `json:"data"`
	Collapse            json.RawMessage     `json:"collapse"`
	Order               json.RawMessage     `json:"order"`
	DatabaseId          int64               `json:"database_id"`
}

// Representation of a single Quandl data bucket, e.g.
//
// [
//	"2018-01-13",
//	14100,
//	12751,
//	13801,
//	13800,
//	13800,
//	13802,
//	37434.61421912
//],
type QuandlBucket struct {
	Date                                   string
	High, Low, Mid, Last, Bid, Ask, Volume float64
}

// Time intervals supported by Quandl
var quandlIntervals = map[string]bool{
	TWOYEAR:    true,
	YEAR:       true,
	SIXMONTH:   true,
	THREEMONTH: true,
	MONTH:      true,
}

const quandlApiVersion = "v3"
const bitfinex = "BITFINEX"
const bitfinexTicker = "BTCUSD"

var quandlBitfinexEndpoint = fmt.Sprintf(quandlEndpoint, quandlApiVersion, bitfinex, bitfinexTicker)

// Given an interval, check its validity and return all Bitfinex data within that interval, as PricePoints
// Currently, we only support Bitfinex data through Quandl, whose finest granularity is one day
//
// TODO: Add direct Bitfinex API to support intervals shorter than 1 month
func PollBitfinexHistorical(interval string) ([]PricePoint, *errors.MyError) {
	interval = strings.ToUpper(interval)
	if !quandlIntervals[interval] {
		return nil, &errors.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: http.StatusBadRequest}
	}

	quandlResponse, err := fetchQuandlResponse(interval)

	if err != nil {
		return nil, err
	}

	return parseQuandlBuckets(quandlResponse.DataSetResponse.Data)
}

// Given the raw 2D Quandl data, convert it to an array of PricePoints
func parseQuandlBuckets(buckets [][]json.RawMessage) ([]PricePoint, *errors.MyError) {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		bucket := new(QuandlBucket)

		err := unmarshalQuandlBucket(val, bucket)

		if err != nil {
			return nil, &errors.MyError{Err: err.Error()}
		}

		timestamp, err := time.Parse("2006-01-02", bucket.Date)

		if err != nil {
			log.Println("Could not parse time from bucket date")
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
func unmarshalQuandlBucket(jsonBucket []json.RawMessage, bucket *QuandlBucket) error {
	err := json.Unmarshal(jsonBucket[0], &bucket.Date)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBucket[3], &bucket.Mid)
	if err != nil {
		return err
	}

	return nil
}

// Given an interval
// 1. Build the GET request
// 2. Fetch the historical data from Quandl
// 3. Return the response if successful, error if not
func fetchQuandlResponse(interval string) (*QuandlResponse, *errors.MyError) {
	requestString, err := buildQuandlRequest(interval)
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

	if response.StatusCode == http.StatusOK {
		quandlResponse := new(QuandlResponse)
		err = json.NewDecoder(response.Body).Decode(quandlResponse)

		if err != nil {
			log.Println("Could not decode Quandl response")
			return nil, &errors.MyError{Err: err.Error()}
		}
		return quandlResponse, nil
	} else {
		log.Println(fmt.Sprintf("Their was an error contacting Quandl with response code %d", response.StatusCode))
		return nil, &errors.MyError{Err: "Quandl API error", ErrorCode: http.StatusInternalServerError}
	}
}

// Given an interval, add the custom GET parameters to the Quandl request
func buildQuandlRequest(interval string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, quandlBitfinexEndpoint, nil)
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

// Similar to CoinDesk, determine the start date for the Quandl response
func getQuandlStartDate(interval string) string {
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
	}

	return startTime.Format(DATELAYOUTSTRING)
}
