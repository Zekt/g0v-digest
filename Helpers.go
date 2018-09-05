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

func Parse(source io.Reader) (SplitedArticle, error) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Parsing html: ", err.Error())
	}

	var digest SplitedArticle

	nodes := doc.Find("h3")
	nodes.Each(func(index int, node *goquery.Selection) {
		h3, err := node.Children().Children().Html()
		imgSrc := node.Next().Children().AttrOr("src", "")
		p, err := node.Next().Next().Html()
		if err != nil {
			log.Println("Reading HTML: ", err.Error())
			return
		}
		digest.Digests = append(digest.Digests, struct{ title, pos, img, content string }{h3, "", imgSrc, p})
	})
	return digest, err
}

func Scrap(source io.Reader, target *SplitedArticle) {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Parsing Medium html: ", err.Error())
		return
	}

	h3s := doc.Find("h3")
	for i, v := range target.Digests {
		prevName := h3s.FilterFunction(func(_ int, sel *goquery.Selection) bool {
			h, err := sel.Children().Children().Html()
			if err != nil {
				log.Println("Failed to read HTML: ", err.Error())
			}
			return h == v.title
		}).Prev().AttrOr("name", "")
		target.Digests[i].pos = prevName
	}
}
