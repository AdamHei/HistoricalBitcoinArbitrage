package handlers

import (
	"encoding/json"
	"github.com/adamhei/historicalapi/errors"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
)

const INTERVAL = "interval"

type AppContext struct {
	Db *mgo.Database
}

func (appcontext *AppContext) Index(responseWriter http.ResponseWriter, request *http.Request) {
	respond(responseWriter, "Welcome to the Bitcoin Historical Data API", nil)
}

func respond(writer http.ResponseWriter, data interface{}, err *errors.MyError) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		log.Println(err.Err)
		errCode := err.ErrorCode
		if errCode == 0 {
			errCode = http.StatusInternalServerError
		}
		http.Error(writer, err.Err, err.ErrorCode)
	} else {
		json.NewEncoder(writer).Encode(data)
	}
}
