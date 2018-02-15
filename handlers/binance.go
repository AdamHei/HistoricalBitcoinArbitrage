package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"github.com/gorilla/mux"
	"net/http"
)

func (appContext *AppContext) BinanceHistorical(responseWriter http.ResponseWriter, req *http.Request) {
	args := mux.Vars(req)
	interval := args[INTERVAL]

	pricePoints, err := datamodels.PollBinanceHistorical(interval)

	if err != nil {
		respond(responseWriter, nil, err)
	} else {
		respond(responseWriter, pricePoints, nil)
	}
}
