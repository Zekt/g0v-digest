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

func NewCampaignRequest(title string) (*http.Request, error) {
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
	campaign.Recipients.ListId = config.ListId
	campaign.Settings.Subject = title
	campaign.Settings.Title = title
	campaign.Settings.FromName = "g0v.news 團隊"
	campaign.Settings.ReplyTo = "g0v.news@ocf.tw"
	campaign.Settings.TempId = config.TempId

	jsonBytes, err := json.Marshal(campaign)
	if err != nil {
		return nil, err
	}
	req, err := NewMailchimpRequest("POST", "/campaigns", jsonBytes)

	return req, err
}

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
