package main

import (
	//"fmt"
	"net/http"
	//"golang.org/x/net/html"
	//"io/ioutil"
	"github.com/ant0ine/go-json-rest/rest"
	"log"
	"strconv"
)

func (t *Trades) GetTrades(w rest.ResponseWriter, r *rest.Request) {
	since, err := strconv.ParseInt(r.FormValue("since"), 0, 64)
	if err != nil {
		since = 0
	}

	trades := t.GetTradesSince(since)
	if trades == nil {
		rest.Error(w, "Failed to get trades", 400)
		return
	}

	w.WriteJson(trades)
}

func GetAPI(w rest.ResponseWriter, r *rest.Request) {
	apis := []string{"/trades"}
	w.WriteJson(apis)
}

func main() {
	trades := NewTrades()

	trades.StartBackground(60)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/trades", trades.GetTrades),
		rest.Get("/", GetAPI),
	)
	if err != nil {
		log.Fatal(err)
	}

	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
