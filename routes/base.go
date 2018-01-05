package routes

import (
	"github.com/adamhei/historicalapi/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

const GET = "GET"

type route struct {
	Name, Method, Path string
	HandlerFunc        http.HandlerFunc
}

func getRoutes(appContext *handlers.AppContext) []route {
	return []route{
		{
			Method:      GET,
			Path:        "/historical/gemini",
			Name:        "Gemini Historical",
			HandlerFunc: appContext.GeminiHistorical,
		},
		{
			Method:      GET,
			Path:        "/historical/gdax/{interval}",
			Name:        "GDAX Historical",
			HandlerFunc: appContext.GdaxHistorical,
		},
	}
}

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
