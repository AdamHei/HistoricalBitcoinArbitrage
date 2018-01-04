package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"net/http"
)

func (appContext *AppContext) GdaxHistorical(responseWriter http.ResponseWriter, req *http.Request) {
	pricePoints, err := datamodels.PollGdaxHistorical(datamodels.SIXMONTH)

	if err != nil {
		respond(responseWriter, nil, err)
	} else {
		respond(responseWriter, pricePoints, nil)
	}
}
