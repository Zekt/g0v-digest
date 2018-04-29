package main

import (
	// "encoding/json"
	// "fmt"
	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func RouteMedium(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		//TODO: fetch all new Medium posts and store to DB.
		/*
			decoder := json.NewDecoder(req.Body)
			var request struct {
				Title       string `json:"title"`
				Url         string `json:"url"`
				PublishedAt string `json:"publishedAt"`
			}
			if err := decoder.Decode(&request); err != nil {
				log.Print("decoding request json: ", err.Error())
				res.WriteHeader(http.StatusBadRequest)
			}
		*/

		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(config.RssUrl)
		if err != nil {
			log.Fatal("parsing rss: ", err.Error())
		}
		for _, v := range feed.Items {
			if strings.Contains(v.Title, "週報") {
				if u, err := url.Parse(v.Link); err != nil {
					log.Println(err.Error())
				} else {
					u.RawQuery = ""
					v.Link = u.String()
				}

				article := Article{
					Title:    v.Title,
					Language: "zh",
					PubTime:  *v.PublishedParsed,
					Url:      v.Link,
					Tags:     v.Categories,
					Html:     v.Content,
				}
				StoreArticle(article)
			}
		}

		res.Write([]byte("Parse done."))
	})
}

func RouteAPI(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		//TODO: return JSON
	})
}
