// Package handlers acts as an interface to the datamodels package and handles all http responses
//
// handler files usually contain one file with a function which
// -parses the client args
// -queries the corresponding datamodel
// -and finally calls respond
package handlers

import (
	"encoding/json"
	"github.com/adamhei/historicalapi/errors"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
)

const INTERVAL = "interval"

// Dependency injection for easy access to the database
//
// Currently obsolete as Gemini is not supported
type AppContext struct {
	Db *mgo.Database
}

// The index endpoint
func (appcontext *AppContext) Index(responseWriter http.ResponseWriter, request *http.Request) {
	respond(responseWriter, "Welcome to the Bitcoin Historical Data API", nil)
}

// respond is the all in one method for writing a response or error back to the client
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
