package main

import (
	//"fmt"
	"golang.org/x/net/html"
	"net/http"
	//"io/ioutil"
	"log"
	"time"
)

const url = "http://www.aktietorget.se/InsiderTransactions.aspx"
const tableName = "ctl00_ctl00_MasterContentBody_InsiderMasterSearchResult_tblInsiderTransactions"

func getAttr(attr []html.Attribute, key string) string {
	for _, a := range attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func getText(n *html.Node) string {
	if n == nil {
		return ""
	}

	data := search(n, func(n *html.Node) bool {
		return n.Type == html.TextNode
	})

	if data == nil {
		return ""
	}

	return data.Data
}

func getTextAndAttr(n *html.Node, attr string) (string, string) {
	if n == nil {
		return "", ""
	}
	return getText(n), getAttr(n.Attr, attr)
}

func search(n *html.Node, f func(*html.Node) bool) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if f(c) {
			return c
		}
		if r := search(c, f); r != nil {
			return r
		}
	}
	return nil
}

func each(n *html.Node, f func(*html.Node)) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f(c)
	}
}

//TODO: return the state holder instead of hiding it inside the closure
func eachgen(n *html.Node, f func(*html.Node) bool) func() *html.Node {
	c := n.FirstChild
	return func() *html.Node {
		for ; c != nil; c = c.NextSibling {
			if f(c) {
				// we need to move the iterator to the next before returning
				rv := c
				c = c.NextSibling
				return rv
			}
		}
		return nil
	}
}

func fetch() (output []Entry, err error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := html.Parse(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	table := search(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "table" && getAttr(n.Attr, "id") == tableName
	})

	rows := search(table, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "tbody"
	})

	// generate all tr
	gentr := eachgen(rows, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "tr"
	})

	output = make([]Entry, 0, 100)

	// for each tr, look for all td's
	for tr := gentr(); tr != nil; tr = gentr() {
		gentd := eachgen(tr, func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == "td"
		})

		date := getText(gentd())
		if date == "" {
			continue
		}
		instr, instrlong := getTextAndAttr(gentd(), "title")
		possession := getText(gentd())
		_ = gentd() // two unused columns
		_ = gentd()
		violation := getText(gentd())
		reason := getText(gentd())
		change := getText(gentd())
		volume := getText(gentd())

		output = append(output, Entry{date, instr, instrlong, possession, violation, reason, change, volume, time.Now().Unix()})
	}

	return output, nil
}

func aktietorget_refresh(storage *Trades) {
	log.Println("Updating aktietorget")
	data, err := fetch()
	if err != nil {
		log.Println("Failed to fetch..")
		return
	}

	storage.NewEntries(data)
}

func aktietorget_background(storage *Trades, refreshAfter time.Duration) chan struct{} {
	quit := make(chan struct{})
	go func() {
		aktietorget_refresh(storage)

		timer := time.NewTicker(time.Second * refreshAfter).C
		for {
			select {
			case <-timer:
				aktietorget_refresh(storage)
			case <-quit:
				return
			}
		}
	}()
	return quit
}
