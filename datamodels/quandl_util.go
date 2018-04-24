package datamodels

import (
	"encoding/json"
	"fmt"
	"github.com/adamhei/historicalapi/errors"
	"log"
	"net/http"
	"time"
)

const quandlApiV3 = "v3"

// Top level response body
type quandlResponse struct {
	DataSetResponse quandlDataSetResponse `json:"dataset"`
}

// Mid level response body
type quandlDataSetResponse struct {
	Id                  int64           `json:"id"`
	DatasetCode         string          `json:"dataset_code"`
	DatabaseCode        string          `json:"database_code"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	RefreshedAt         string          `json:"refreshed_at"`
	NewestAvailableDate string          `json:"newest_available_date"`
	OldestAvailableDate string          `json:"oldest_available_date"`
	ColumnNames         []string        `json:"column_names"`
	Frequency           string          `json:"frequency"`
	Type                string          `json:"type"`
	Premium             bool            `json:"premium"`
	Limit               json.RawMessage `json:"limit"`
	Transform           json.RawMessage `json:"transform"`
	ColumnIndex         json.RawMessage `json:"column_index"`
	StartDate           string          `json:"start_date"`
	EndDate             string          `json:"end_date"`
	// The only relevant part
	Data       [][]json.RawMessage `json:"data"`
	Collapse   json.RawMessage     `json:"collapse"`
	Order      json.RawMessage     `json:"order"`
	DatabaseId int64               `json:"database_id"`
}

// Time intervals supported by Quandl
var quandlIntervals = map[string]bool{
	TWOYEAR:    true,
	YEAR:       true,
	SIXMONTH:   true,
	THREEMONTH: true,
	MONTH:      true,
}

// Given an interval
// 1. Build the GET request
// 2. Fetch the historical data from Quandl
// 3. Return the response if successful, error if not
func fetchQuandlResponse(requestString string) (*quandlResponse, *errors.MyError) {
	response, err := http.Get(requestString)

	defer response.Body.Close()

	log.Println(fmt.Sprintf("Querying %s", requestString))

	if err != nil {
		log.Println(fmt.Sprintf("Could not reach %s", requestString))
		return nil, &errors.MyError{Err: err.Error()}
	}

	if response.StatusCode == http.StatusOK {
		quandlResponse := new(quandlResponse)
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
