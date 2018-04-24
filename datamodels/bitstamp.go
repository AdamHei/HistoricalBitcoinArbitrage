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

const bitstamp = "BITSTAMP"
const qBitstampTemplate = "https://www.quandl.com/api/%s/datasets/%s/USD.json"

var qBitstampEndpoint = fmt.Sprintf(qBitstampTemplate, quandlApiV3, bitstamp)

type qBitstampBudcket struct {
	Date                                    string
	High, Low, Last, Bid, Ask, Volume, VWAP float64
}

func PollBitstampHistorical(interval string) ([]PricePoint, *errors.MyError) {
	interval = strings.ToUpper(interval)
	if !quandlIntervals[interval] {
		return nil, &errors.MyError{Err: fmt.Sprintf("Please provide a valid interval; %s is invalid", interval), ErrorCode: http.StatusBadRequest}
	}

	requestString, err := buildQBitstampRequest(interval)
	if err != nil {
		return nil, &errors.MyError{Err: err.Error()}
	}

	quandlReponse, myErr := fetchQuandlResponse(requestString)
	if myErr != nil {
		return nil, myErr
	}

	return parseQBitstampBuckets(quandlReponse.DataSetResponse.Data)
}

func parseQBitstampBuckets(buckets [][]json.RawMessage) ([]PricePoint, *errors.MyError) {
	pricePoints := make([]PricePoint, len(buckets))

	for index, val := range buckets {
		bucket, err := unmarshalQBitstampBucket(val)
		if err != nil {
			return nil, &errors.MyError{Err: err.Error()}
		}

		timestamp, err := time.Parse(DATELAYOUTSTRING, bucket.Date)
		if err != nil {
			log.Println("Could not parse time from Bitstamp bucket")
			return nil, &errors.MyError{Err: "Failure to parse Quandl response", ErrorCode: http.StatusInternalServerError}
		}

		price := strconv.FormatFloat(bucket.VWAP, 'f', -1, 64)

		pricePoints[index] = PricePoint{Timestamp: timestamp.Unix(), Price: price}
	}

	return pricePoints, nil
}

func unmarshalQBitstampBucket(jsonBucket []json.RawMessage) (*qBitstampBudcket, error) {
	bucket := new(qBitstampBudcket)

	err := json.Unmarshal(jsonBucket[0], &bucket.Date)
	if err != nil {
		log.Println("Failed to parse Bitstamp bucket date")
		return nil, err
	}

	err = json.Unmarshal(jsonBucket[7], &bucket.VWAP)
	if err != nil {
		log.Println("Failed to parse Bitstamp bucket VWAP")
		return nil, err
	}
	return bucket, nil
}

func buildQBitstampRequest(interval string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, qBitstampEndpoint, nil)
	if err != nil {
		log.Println("Could not build Quandl Bitstamp URL")
		return EMPTYSTRING, err
	}

	query := request.URL.Query()
	query.Add("api_key", quandlApiKey)
	query.Add("start_date", getQuandlStartDate(interval))

	request.URL.RawQuery = query.Encode()
	return request.URL.String(), nil
}
