package main

import "net/http"

func StringInSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}

func RequestMailchimp(Methods string, EntryPoint string) (*http.Request, error) { //unused
	req, err := http.NewRequest(Methods, config.ApiUrl+EntryPoint, nil)
	if err == nil {
		req.Header.Set("Authorization", "Basic "+config.ApiKey)
		req.Header.Set("Content-Type", "application/json")
	}
	return req, err
}
