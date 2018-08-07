package main

import (
	"bytes"
	"encoding/json"
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

func NewCampaignRequest(title string, listId string) (*http.Request, error) {
	type Campaign struct {
		Type   string `json:"type"`
		ListId string `json:"recipients>list_id"`
		Title  string `json:"settings>title"`
	}
	camp := Campaign{
		Type:   "regular",
		ListId: listId,
		Title:  title,
	}

	jsonBytes, err := json.Marshal(camp)
	if err != nil {
		return nil, err
	}
	req, err := NewMailchimpRequest("POST", "/campaigns", jsonBytes)

	return req, err
}
