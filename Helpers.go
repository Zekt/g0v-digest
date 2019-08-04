package main

import (
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
)

func StringInSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}

func NewMailchimpRequest(methods string, resource string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(methods, config.ApiUrl+resource, bytes.NewReader(body))
	if err == nil {
		req.Header.Set("Authorization", "Basic "+config.ApiKey)
		req.Header.Set("content-type", "application/json")
		req.SetBasicAuth("anystring", config.ApiKey)
	}
	return req, err
}

func NewCampaignRequest(title, lang string) (*http.Request, error) {
	var campaign struct {
		Type       string `json:"type"`
		Recipients struct {
			ListId string `json:"list_id"`
		} `json:"recipients"`
		Settings struct {
			Subject  string `json:"subject_line"`
			Title    string `json:"title"`
			FromName string `json:"from_name"`
			ReplyTo  string `json:"reply_to"`
			TempId   int    `json:"template_id"`
		} `json:"settings"`
	}

	campaign.Type = "regular"
	campaign.Settings.Subject = title
	campaign.Settings.Title = title
	campaign.Settings.FromName = "g0v.news"
	campaign.Settings.ReplyTo = "g0v.news@ocf.tw"
	if lang == "en" {
		campaign.Settings.TempId = config.TempIdEn
		campaign.Recipients.ListId = config.ListIdEn
	} else {
		campaign.Settings.TempId = config.TempId
		campaign.Recipients.ListId = config.ListId
	}

	jsonBytes, err := json.Marshal(campaign)
	if err != nil {
		return nil, err
	}
	req, err := NewMailchimpRequest("POST", "/campaigns", jsonBytes)

	return req, err
}

//Main parser from HTML to mailchimp template.
func Parse(source io.Reader, lang string) (SplitedArticle, error) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Parsing html: ", err.Error())
	}

	var digest SplitedArticle

	var nodes *goquery.Selection
	if lang == "en" {
		nodes = doc.Find("h4").Slice(1, 5)
	} else {
		nodes = doc.Find("h3")
	}
	nodes.Each(func(index int, node *goquery.Selection) {
		title, err := node.Children().Children().Html()
		if title == "" || err != nil {
			title, err = node.Children().Html()
			if err != nil {
				log.Println("Reading "+lang+" RSS HTML: ", err.Error())
				return
			}
		}
		imgSrc := node.Next().Children().AttrOr("src", "")
		p, err := node.Next().Next().Html()
		if err != nil {
			log.Println("Reading "+lang+" RSS HTML: ", err.Error())
			return
		}
		digest.Digests = append(digest.Digests, struct{ title, pos, img, content string }{title, "", imgSrc, p})
	})
	return digest, err
}

//DEPRECATED: Scrap section id from Medium Source.
func Scrap(source io.Reader, target *SplitedArticle, lang string) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Parsing "+lang+" Medium html: ", err.Error())
		return
	}

	var titles *goquery.Selection
	if lang == "en" {
		titles = doc.Find("h4").Slice(0, 4)
	} else {
		titles = doc.Find("h3")
	}
	for i, v := range target.Digests {
		prevName := titles.FilterFunction(func(_ int, sel *goquery.Selection) bool {
			h, err := sel.Children().Children().Html()
			if h == "" || err != nil {
				h, err = sel.Children().Html()
				if err != nil {
					log.Println("Failed to read "+lang+" HTML Title: ", err.Error())
				}
			}
			return h == v.title
		}).Prev().AttrOr("name", "")
		target.Digests[i].pos = prevName
	}
}

func RemovePixel(source io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Error parsing html: ", err.Error())
	}

	nodes := doc.Find("img")
	nodes.Each(func(index int, node *goquery.Selection) {
		val, exist := node.Attr("width")
		if exist == true && val == "1" {
			node.Remove()
		}
	})

	return doc.Html()
}
