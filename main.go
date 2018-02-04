package main

import (
	"github.com/adamhei/historicalapi/handlers"
	"github.com/adamhei/historicalapi/routes"
	"github.com/adamhei/historicaldata/trademodels"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
	"time"
)

func main() {
	mgoDialInfo := &mgo.DialInfo{
		Addrs:    []string{trademodels.DbUrl},
		Timeout:  1 * time.Hour,
		Database: trademodels.AUTHDB,
		Username: trademodels.USERNAME,
		Password: trademodels.PASSWORD,
	}
	sesh, err := mgo.DialWithInfo(mgoDialInfo)
	defer sesh.Close()

	if err != nil {
		log.Println("Could not connect to DB")
		panic(err)
	}

	db := sesh.DB(trademodels.DbName)

	appContext := &handlers.AppContext{Db: db}
	router := routes.NewRouter(appContext)

	log.Fatal(http.ListenAndServe(":80", router))
}
