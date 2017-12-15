package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"net/http"
)

func (appContext *AppContext) Historical(w http.ResponseWriter, r *http.Request) {
	results, err := datamodels.GetHistorical(appContext.Db, datamodels.TWOYEAR)
	respond(w, results, err)
}
