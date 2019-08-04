package main

import (
	"encoding/xml"
	"fmt"
	// "bytes"
	"encoding/json"
	xj "github.com/basgys/goxml2json"
	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func RouteWordpress(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		resMessage := "Done parsing from Wordpress."
		client := &http.Client{}
		reqJson := make(map[string]interface{})

		reqWordpress, err := http.NewRequest("GET", config.WordpressUrl, nil)
		if err != nil {
			log.Println("Building request to Wordpress failed: ", err.Error())
			return
		}
		q := reqWordpress.URL.Query()
		q.Add("fileds", "ID,date,title,URL,content,excerpt,tags")
		q.Add("tag", "DIGEST")
		reqWordpress.URL.RawQuery = q.Encode()

		resWordpress, err := client.Do(reqWordpress)
		if err != nil {
			log.Println("Making request to Wordpress failed: ", err.Error())
			return
		}

		resBody, err := ioutil.ReadAll(resWordpress.Body)
		if err != nil {
			log.Println("Reading Wordpress response failed: ", err.Error())
			return
		}

		err = json.Unmarshal(resBody, &reqJson)
		if err != nil {
			log.Println("Parsing Wordpress response in JSON failed: ", err.Error())
			return
		}

		if val, ok := reqJson["posts"]; ok {
			for _, v := range val.([]interface{}) {
				v := v.(map[string]interface{})
				pub, err := time.Parse(time.RFC3339, v["date"].(string))
				if err != nil {
					log.Println("Error parsing time: ", err.Error())
				}
				article := Article{
					Title:   v["title"].(string),
					PubTime: pub,
					Url:     v["URL"].(string),
					Html:    v["content"].(string),
				}
				if val, ok := v["tags"].(map[string]interface{}); ok {
					for k, _ := range val {
						article.Tags = append(article.Tags, k)
					}
					if StringInSlice("中文", article.Tags) {
						article.Language = "zh"
					}
				}
				StoreArticle(article, func() {
					client := &http.Client{}
					req, err := http.NewRequest(
						"PUT",
						fmt.Sprintf("http://%s:%d/mailchimp?lang=%s", config.Server, config.Port, article.Language),
						nil,
					)
					if err != nil {
						log.Println("Building request to Mailchimp on update from Medium: ", err.Error())
						return
					}
					_, err = client.Do(req)
					if err != nil {
						log.Println("Sending request to Mailchimp on update from Wordpress: ", err.Error())
						return
					}
					resMessage += "\nMailchimp content updated."
				})
			}
		} else {
			log.Println("No \"posts\" field found.")
		}

		res.Write([]byte(resMessage))
	})
}

func RouteMedium(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		resMessage := "Done parsing from Medium."
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

				StoreArticle(article, func() {
					client := &http.Client{}
					req, err := http.NewRequest(
						"PUT",
						fmt.Sprintf("http://%s:%d/mailchimp?lang=%s", config.Server, config.Port, article.Language),
						nil,
					)
					if err != nil {
						log.Println("Building request to Mailchimp on update from Medium: ", err.Error())
						return
					}
					_, err = client.Do(req)
					if err != nil {
						log.Println("Sending request to Mailchimp on update from Medium: ", err.Error())
						return
					}
					resMessage += "\nMailchimp content updated."
				})
			}
		}

		res.Write([]byte(resMessage))
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
		var rss []byte
		var err error
		lang := req.FormValue("lang")
		if lang == "en" {
			rss, err = GetNewestXML("en")
		} else {
			rss, err = GetNewestXML("zh")
		}
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusBadRequest)

		}
		res.Header().Set("Content-Type", "application/xml")
		res.Write(append([]byte(xml.Header), rss...))
	})

	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
		//TODO: return JSON, error not handled.
		client := &http.Client{}
		lang := req.FormValue("lang")
		url := fmt.Sprintf("http://%s:%d/api/line?lang=%s", config.Server, config.Port, lang)
		reqLocal, _ := http.NewRequest("GET", url, nil)
		resLocal, _ := client.Do(reqLocal)
		bytes, err := xj.Convert(resLocal.Body)
		if err != nil {
			log.Println("Converting to json failed: ", err.Error())
		}
		res.Write(bytes.Bytes())
	})
}

func RouteMailchimp(sub *mux.Router) {
	sub.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		lang := req.FormValue("lang")
		client := &http.Client{}
		var reqJson struct {
			Template struct {
				ID       int               `json:"id"`
				Sections map[string]string `json:"sections"`
			} `json:"template"`
		}

		sections := make(map[string]string)

		article, err := GetArticle(lang, 0)
		if err != nil {
			log.Println("Querying latest article: ", err.Error())
			return
		}
		parsedArticle, err := Parse(strings.NewReader(article.Html), lang)
		if err != nil {
			log.Println("Parsing error")
			return
		}

		/*
			reqMedium, err := http.NewRequest("GET", article.Url, nil)
			if err != nil {
				log.Println("Building request to Medium failed: ", err.Error())
				return
			}

			resMedium, err := client.Do(reqMedium)
			if err != nil {
				log.Println("Making request to Medium failed: ", err.Error())
				return
			}

			Scrap(resMedium.Body, &parsedArticle, lang)
		*/

		t := article.PubTime
		sections["date"] = fmt.Sprintf("%s %d WEEKLY NEWS", t.Month().String()[:3], t.Day())
		if lang == "en" {
			sections["link"] = fmt.Sprintf("<a href=\"%s\">%s</a>", article.Url, "This Week on Civic Tech")
		} else {
			sections["link"] = fmt.Sprintf("<a href=\"%s\">%s</a>", article.Url, "一 週 公 民 科 技 焦 點")
		}
		for i, v := range parsedArticle.Digests {
			s := strconv.Itoa(i + 1)
			sections["title"+s] = fmt.Sprintf("<a href=\"%s#%s\">%s</a>", article.Url, v.pos, v.title)
			sections["content"+s] = v.content
		}

		if lang == "en" {
			reqJson.Template.ID = config.TempIdEn
		} else {
			reqJson.Template.ID = config.TempId
		}
		reqJson.Template.Sections = sections

		jsonBytes, err := json.Marshal(reqJson)
		if err != nil {
			log.Println("Marshalling json: ", err.Error())
			return
		}

		// POST to creat a new campaign and get that campaign ID.
		reqCamp, err := NewCampaignRequest(article.Title, lang)
		if err != nil {
			log.Println("Building request to create new campaign: ", err.Error())
			return
		}
		resCamp, err := client.Do(reqCamp)
		if err != nil {
			log.Println("Making request to create new campaign: ", err.Error())
			return
		}
		var camp struct {
			ID string `json:"id"`
		}
		resBody, err := ioutil.ReadAll(resCamp.Body)
		if err != nil {
			log.Println("Reading Mailchimp response to creating new campaign: ", err.Error())
			return
		}

		err = json.Unmarshal(resBody, &camp)
		if err != nil {
			log.Println("Parsing Mailchimp response in JSON: ", err.Error())
			return
		}

		// Update that Mailchimp campaign based on ID.
		reqMC, err := NewMailchimpRequest("PUT", "/campaigns/"+camp.ID+"/content", jsonBytes)
		if err != nil {
			log.Println("Making a request to Mailchimp: ", err.Error())
			return
		}
		resM, err := client.Do(reqMC)
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

		res.Write([]byte("Mailchimp content updated."))
	})
}
