package handlers

import (
	"github.com/adamhei/historicalapi/datamodels"
	"net/http"
)

func (appContext *AppContext) GeminiHistorical(w http.ResponseWriter, r *http.Request) {
	results, err := datamodels.QueryGeminiHistorical(appContext.Db, datamodels.TWOYEAR)
	respond(w, results, err)
}
