package handlers

import (
	"encoding/json"
	"github.com/adamhei/historicalapi/errorhandling"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
)

const INTERVAL = "interval"

type AppContext struct {
	Db *mgo.Database
}

func respond(writer http.ResponseWriter, data interface{}, err *errorhandling.MyError) {
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