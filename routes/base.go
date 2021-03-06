// Package routes provides a directory of how to use the Historical API
package routes

import (
	"github.com/adamhei/historicalapi/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

// Wrapper for consistent route definition
type route struct {
	Name, Method, Path string
	HandlerFunc        http.HandlerFunc
}

// NewRouter constructs and returns a mux Router with all routes in the API
func NewRouter(appContext *handlers.AppContext) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, r := range getRoutes(appContext) {
		router.Methods(r.Method).
			Path(r.Path).
			Name(r.Name).
			HandlerFunc(r.HandlerFunc)
	}

	return router
}

func getRoutes(appContext *handlers.AppContext) []route {
	return []route{
		{
			Method:      http.MethodGet,
			Path:        "/",
			Name:        "Index page",
			HandlerFunc: appContext.Index,
		},
		//{
		//	Method:      http.MethodGet,
		//	Path:        "/historical/gemini",
		//	Name:        "Gemini Historical",
		//	HandlerFunc: appContext.GeminiHistorical,
		//},
		{
			Method:      http.MethodGet,
			Path:        "/historical/gdax/{interval}",
			Name:        "GDAX Historical",
			HandlerFunc: appContext.GdaxHistorical,
		},
		{
			Method:      http.MethodGet,
			Path:        "/historical/kraken/{interval}",
			Name:        "Kraken Historical",
			HandlerFunc: appContext.KrakenHistorical,
		},
		{
			Method:      http.MethodGet,
			Path:        "/historical/bitfinex/{interval}",
			Name:        "Bitfinex Historical",
			HandlerFunc: appContext.BitfinexHistorical,
		},
		{
			Method:      http.MethodGet,
			Path:        "/historical/index/{interval}",
			Name:        "Index Price (from CoinDesk)",
			HandlerFunc: appContext.CoinDeskHistorical,
		},
		{
			Method:      http.MethodGet,
			Path:        "/historical/binance/{interval}",
			Name:        "Binance Historical",
			HandlerFunc: appContext.BinanceHistorical,
		},
		{
			Method:      http.MethodGet,
			Path:        "/historical/bitstamp/{interval}",
			Name:        "Bitstamp Historical",
			HandlerFunc: appContext.BitstampHistorical,
		},
	}
}
