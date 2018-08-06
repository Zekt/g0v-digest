package main

import (
	"encoding/xml"
	// "fmt"
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
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
	sub.PathPrefix("/line/tick").Methods("PUT").HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		err := UpdateXmlUuid()
		if err != nil {
			log.Println("Updating id error:", err.Error())
			res.Write([]byte("Update failed!"))
		} else {
			res.Write([]byte("Update done."))
		}
	})
	sub.HandleFunc("/line", func(res http.ResponseWriter, req *http.Request) {
		// Retuen XML for LINE Today
		rss, err := GetNewestXML()
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusBadRequest)

		}
		res.Header().Set("Content-Type", "application/xml")
		res.Write(append([]byte(xml.Header), rss...))
	})

	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
		//TODO: return JSON
	})
}

func RouteMailchimp(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		client := &http.Client{}
		var reqJson struct {
			html string
		}
		html, err := GetArticle()
		if err != nil {
			log.Println("Querying latest article: ", err.Error())
			return
		}
		reqJson.html = html
		jsonM, err := json.Marshal(reqJson)
		if err != nil {
			log.Println("Marshalling json: ", err.Error())
		}
		reqM, err := http.NewRequest(
			"PUT",
			config.ApiUrl+"/campaigns/"+config.CampId+"/content",
			bytes.NewReader(jsonM),
		)
		if err != nil {
			log.Println("Making a request to Mailchimp: ", err.Error())
			return
		}
		reqM.Header.Set("Authorization", "Basic "+config.ApiKey)
		reqM.Header.Set("Content-Type", "application/json")
		resM, err := client.Do(reqM)
		if err != nil {
			log.Println("Sending Mailchimp request:", err.Error())
			return
		}
		if resM.StatusCode != http.StatusOK {
			errMsg, err := ioutil.ReadAll(resM.Body)
			if err != nil {
				log.Println("Error reading response body: ", err.Error())
			} else {
				log.Println("Getting Mailchimp status: ", resM.Status)
				log.Println("Getting Mailchimp response: \n", string(errMsg))
			}
			res.Write([]byte("Something went wrong! Writting Mailchimp content failed."))
			return
		}
		/*
			text, err := ioutil.ReadAll(resM.Body)
			if err != nil {
				log.Println("Reading Mailchimp response", err.Error())
				return
			}
			var response struct {
				Html string `json:html`
			}
			err = json.Unmarshal(text, &response)
			log.Println(string(response.Html))
			if err != nil {
				log.Println("Parsing Mailchimp json", err.Error())
				return
			}
		*/

		res.Write([]byte("Mailchimp content updated."))
	})
}
