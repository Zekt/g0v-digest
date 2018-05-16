package main

import (
	"encoding/xml"
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
			if StringInSlice("digest", v.Categories) || strings.Contains(v.Title, "週報") {
				if u, err := url.Parse(v.Link); err != nil {
					log.Println(err.Error())
				} else {
					u.RawQuery = "" // Remove query string appended in Medium RSS.
					v.Link = u.String()
				}

				article := Article{
					Title:   v.Title,
					PubTime: *v.PublishedParsed,
					Url:     v.Link,
					Tags:    v.Categories,
					Html:    v.Content,
				}

				if StringInSlice("zh", v.Categories) {
					article.Language = "zh"
				} else if StringInSlice("en", v.Categories) {
					article.Language = "en"
				}

				StoreArticle(article)
			}
		}

		res.Write([]byte("Parse done."))
	})
}

func RouteAPI(sub *mux.Router) {
	sub.HandleFunc("/line", func(res http.ResponseWriter, req *http.Request) {
		//TODO: return XML for LINE Today
		rss, err := GetNewestXML()
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		res.Header().Set("Content-Type", "application/xml")
		res.Write(append([]byte(xml.Header), rss...))
	})

	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		//TODO: return JSON
	})
}
