package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CoinDeskResponse struct {
	BPI        map[string]float64 `json:"bpi"`
	Disclaimer string             `json:"disclaimer"`
	Time       CoinDeskTimeData   `json:"time"`
}

type CoinDeskTimeData struct {
	Updated    string `json:"updated"`
	UpdatedIso string `json:"updatedISO"`
}

// Time intervals supported by Coin Desk, since their finest granularity is 1 day
var coinDeskIntervals = map[string]bool{
	TWOYEAR:    true,
	YEAR:       true,
	SIXMONTH:   true,
	THREEMONTH: true,
	MONTH:      true,
}

const coinDeskApiVersion = "v1"
const coinDeskEndpoint = "https://api.coindesk.com/%s/bpi/historical/open.json"

var coinDeskHistoricalEndpoint = fmt.Sprintf(coinDeskEndpoint, coinDeskApiVersion)

// Given an interval, check its validity and return all CoinDesk Bitcoin Price Index data within that interval, as PricePoints
// Currently, we only support 1 month as the shortest lookback period, since the finest granularity of data is 1 day
//
// TODO: Add support for shorter, finer lookbacks (coinmarketcap?)
func PollCoinDeskHistorical(interval string) ([]PricePoint, *errorhandling.MyError) {
	interval = strings.ToUpper(interval)
	if !coinDeskIntervals[interval] {
		return nil, &errorhandling.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: http.StatusBadRequest}
	}

	coinDeskResponse, err := fetchCoinDeskResponse(interval)

	if err != nil {
		return nil, err
	}

	return parseCoinDeskBuckets(coinDeskResponse.BPI)
}

// Given the 2D date -> price response from CoinDesk, convert the data to PricePoints
func parseCoinDeskBuckets(buckets map[string]float64) ([]PricePoint, *errorhandling.MyError) {
	pricePoints := make([]PricePoint, len(buckets))

	index := 0
	for date, price := range buckets {
		timestamp, err := time.Parse(DATELAYOUTSTRING, date)

		if err != nil {
			log.Println("Could not parse CoinDesk date")
			return nil, &errorhandling.MyError{Err: "Could not properly parse CoinDesk response", ErrorCode: http.StatusInternalServerError}
		}

		priceString := strconv.FormatFloat(price, 'f', -1, 64)

		pricePoints[index] = PricePoint{Timestamp: timestamp.Unix(), Price: priceString}
		index++
	}

	// Sort because iteration over a map doesn't preserve insertion order
	sort.Slice(pricePoints, func(i, j int) bool {
		return pricePoints[i].Timestamp >= pricePoints[j].Timestamp
	})

	return pricePoints, nil
}

// Given an interval:
// 1. Build the GET request
// 2. Fetch the historical index data from CoinDesk
// 3. Return the response if successful, error if not
func fetchCoinDeskResponse(interval string) (*CoinDeskResponse, *errorhandling.MyError) {
	requestString, err := buildCoinDeskRequest(interval)
	if err != nil {
		return nil, &errorhandling.MyError{Err: err.Error()}
	}

	response, err := http.Get(requestString)
	defer response.Body.Close()

	log.Println(fmt.Sprintf("Querying %s", requestString))

	if err != nil {
		log.Println(fmt.Sprintf("Could not reach %s", requestString))
		return nil, &errorhandling.MyError{Err: err.Error()}
	}

	if response.StatusCode == http.StatusOK {
		coinDeskResponse := new(CoinDeskResponse)
		err = json.NewDecoder(response.Body).Decode(coinDeskResponse)

		if err != nil {
			log.Println("Could not decode CoinDesk response")
			return nil, &errorhandling.MyError{Err: err.Error()}
		}
		return coinDeskResponse, nil
	} else {
		log.Println(fmt.Sprintf("There was an error contacting Coin Desk with response code %d", response.StatusCode))
		return nil, &errorhandling.MyError{Err: "Coin Desk API error", ErrorCode: http.StatusInternalServerError}
	}
}

// Given an interval, determine the start and end date for the CoinDesk request and construct it
func buildCoinDeskRequest(interval string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, coinDeskHistoricalEndpoint, nil)
	if err != nil {
		log.Println("Could not build CoinDesk URL")
		return EMPTYSTRING, err
	}

	query := request.URL.Query()

	startDate := getCoinDeskStartDate(interval)
	endDate := time.Now().Format("2006-01-02")

	query.Add("start", startDate)
	query.Add("end", endDate)

	request.URL.RawQuery = query.Encode()
	return request.URL.String(), nil
}

// Similar to Bitfinex, only support TWOYEAR through MONTH
func getCoinDeskStartDate(interval string) string {
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

	return startTime.Format("2006-01-02")
}
