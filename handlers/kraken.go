package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"github.com/gorilla/mux"
	"net/http"
)

func (appContext *AppContext) KrakenHistorical(responseWriter http.ResponseWriter, request *http.Request) {
	args := mux.Vars(request)
	interval := args["interval"]

	pricePoints, err := datamodels.PollKrakenHistorical(interval)

	if err != nil {
		respond(responseWriter, nil, err)
	} else {
		respond(responseWriter, pricePoints, nil)
	}
}
