package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"github.com/gorilla/mux"
	"net/http"
)

func (appContext *AppContext) GdaxHistorical(responseWriter http.ResponseWriter, req *http.Request) {
	args := mux.Vars(req)
	interval := args["interval"]

	pricePoints, err := datamodels.PollGdaxHistorical(interval)

	if err != nil {
		respond(responseWriter, nil, err)
	} else {
		respond(responseWriter, pricePoints, nil)
	}
}
