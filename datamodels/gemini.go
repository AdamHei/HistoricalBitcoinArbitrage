package datamodels

import (
	"fmt"
	"github.com/adamhei/historicalapi/errorhandling"
	"github.com/adamhei/historicaldata/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"time"
)

type Interval string

// TODO Return array of PricePoints
func QueryGeminiHistorical(db *mgo.Database, interval Interval) ([]models.GeminiOrder, *errorhandling.MyError) {
	coll := db.C(models.GeminiCollection)

	startTimeMs := getStartTimeMs(interval)

	log.Println(fmt.Sprintf("Searching for Gemini trades since %s", time.Unix(0, startTimeMs*int64(time.Millisecond))))

	query := coll.Find(bson.M{"timestampms": bson.M{"$gte": startTimeMs}})

	count, err := query.Count()
	if err != nil {
		log.Println("Could not count the number of Gemini trades")
		panic(err)
	}
	log.Println(fmt.Sprintf("Found %d trades", count))

	results := make([]models.GeminiOrder, 0)
	err = query.All(&results)

	if err != nil {
		return nil, &errorhandling.MyError{Err: err.Error(), ErrorCode: 500}
	} else {
		return results, nil
	}
}

// Return the time in milliseconds that is one "interval" from now
func getStartTimeMs(interval Interval) int64 {
	startTime := roundTime(time.Now())

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
		startTime = startTime.AddDate(0, 0, -4)
	case DAY:
		startTime = startTime.AddDate(0, 0, -1)
	case TWELVEHOUR:
		startTime = startTime.Add(-12 * time.Hour)
	case SIXHOUR:
		startTime = startTime.Add(-6 * time.Hour)
	case HOUR:
		startTime = startTime.Add(-1 * time.Hour)
	case THIRTYMINUTE:
		startTime = startTime.Add(-30 * time.Minute)
	}

	return startTime.Unix() * 1000
}
