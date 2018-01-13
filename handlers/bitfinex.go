package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"github.com/gorilla/mux"
	"net/http"
)

func (appcontext *AppContext) BitfinexHistorical(responseWriter http.ResponseWriter, request *http.Request) {
	args := mux.Vars(request)
	interval := args[INTERVAL]

	pricePoints, err := datamodels.PollBitfinexHistorical(interval)

	if err != nil {
		respond(responseWriter, nil, err)
	} else {
		respond(responseWriter, pricePoints, nil)
	}
}
