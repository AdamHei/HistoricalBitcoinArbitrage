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

const (
	TWOYEAR      Interval = "TWOYEAR"
	YEAR         Interval = "YEAR"
	SIXMONTH     Interval = "SIXMONTH"
	THREEMONTH   Interval = "THREEMONTH"
	MONTH        Interval = "MONTH"
	WEEK         Interval = "WEEK"
	DAY          Interval = "DAY"
	TWELVEHOUR   Interval = "TWELVEHOUR"
	SIXHOUR      Interval = "SIXHOUR"
	HOUR         Interval = "HOUR"
	THIRTYMINUTE Interval = "THIRTYMINUTE"
)

func GetHistorical(db *mgo.Database, interval Interval) ([]models.GeminiOrder, *errorhandling.MyError) {
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
	timestamp := time.Now()

	switch interval {
	case TWOYEAR:
		timestamp = timestamp.AddDate(-2, 0, 0)
	case YEAR:
		timestamp = timestamp.AddDate(-1, 0, 0)
	case SIXMONTH:
		timestamp = timestamp.AddDate(0, -6, 0)
	case THREEMONTH:
		timestamp = timestamp.AddDate(0, -3, 0)
	case MONTH:
		timestamp = timestamp.AddDate(0, -1, 0)
	case WEEK:
		timestamp = timestamp.AddDate(0, 0, -4)
	case DAY:
		timestamp = timestamp.AddDate(0, 0, -1)
	case TWELVEHOUR:
		timestamp = timestamp.Add(-12 * time.Hour)
	case SIXHOUR:
		timestamp = timestamp.Add(-6 * time.Hour)
	case HOUR:
		timestamp = timestamp.Add(-1 * time.Hour)
	case THIRTYMINUTE:
		timestamp = timestamp.Add(-30 * time.Minute)
	}

	return timestamp.Unix() * 1000
}
