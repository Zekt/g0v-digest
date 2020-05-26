package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID           int    `json:"id"`
				IsBot        bool   `json:"is_bot"`
				FirstName    string `json:"first_name"`
				Username     string `json:"username"`
				LanguageCode string `json:"language_code"`
			} `json:"from"`
			Chat struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date     int    `json:"date"`
			Text     string `json:"text"`
			Entities []struct {
				Offset int    `json:"offset"`
				Length int    `json:"length"`
				Type   string `json:"type"`
			} `json:"entities"`
		} `json:"message"`
	} `json:"result"`
}

var Url = "https://api.telegram.org/bot675702975:AAH8sEPLFgfVe50hXkFqWQXVYAHelXE-7qc/"
var Client = &http.Client{}
var WhiteList = []int{135263559,130325609}

func main() {

	count := 0
	offset := -1
	for {
		req, err := http.NewRequest("GET", Url+"getUpdates?offset="+strconv.Itoa(offset), nil)
		if err != nil {
			log.Println("Error building request: ", err.Error())
			continue
		}
		response, err := ReadRes(req)
		if err != nil {
			log.Println("Error reading response: ", err.Error())
			continue
		}
		for _, v := range response.Result {
			offset = v.UpdateID + 1
			for _, id := range WhiteList {
				if v.Message.From.ID == id {
					goto pass
				}
			}
			log.Println("Not in whitelist.")
			_, err = sendMessage(v.Message.Chat.ID, "You are not authorized.", v.Message.MessageID)
			if err != nil {
				log.Println("Error replying message: ", err.Error())
			}
			continue
		pass:
			if v.Message.MessageID > count {
				count = v.Message.MessageID
				if v.Message.Text == "/update_zh" || v.Message.Text == "/update_zh@g0vDigestBot" {
					if err := Update("zh"); err != nil {
						log.Println("Error updating campaign: ", err.Error())
						_, err := sendMessage(v.Message.Chat.ID, "Chinese campaign update failed.", v.Message.MessageID)
						if err != nil {
							log.Println("Error replying message: ", err.Error())
						}
					} else {
						fmt.Println("Update ZH!")
						_, err := sendMessage(v.Message.Chat.ID, "Chinese campaign updated.", v.Message.MessageID)
						if err != nil {
							log.Println("Error replying message: ", err.Error())
						}
					}
				} else if v.Message.Text == "/update_en" || v.Message.Text == "/update_en@g0vDigestBot" {
					if err := Update("en"); err != nil {
						log.Println("Error updating campaign: ", err.Error())
						_, err := sendMessage(v.Message.Chat.ID, "English campaign update failed.", v.Message.MessageID)
						if err != nil {
							log.Println("Error replying message: ", err.Error())
						}
					} else {
						fmt.Println("Update EN!")
						_, err := sendMessage(v.Message.Chat.ID, "English campaign updated.", v.Message.MessageID)
						if err != nil {
							log.Println("Error replying message: ", err.Error())
						}
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func Update(lang string) error {
	port := 8080
	client := &http.Client{}
	url := "http://localhost"
	req, err := http.NewRequest("PUT", url+":"+strconv.Itoa(port)+"/mailchimp?lang="+lang, nil)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error building request to Mailchimp: ", err.Error()))
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error sending request to local server: ", err.Error()))
	}
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error reading response from local server: ", err.Error()))
	}
	fmt.Println(string(resBytes))
	return nil
}

func ReadRes(req *http.Request) (*Response, error) {
	res, err := Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Sending request to Telegram: ", err.Error()))
	}
	resBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error reading response from Telegram: ", err.Error()))
	}
	var response Response
	err = json.Unmarshal(resBytes, &response)
	if len(response.Result) > 0 {
		log.Println(string(resBytes))
	}

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error unmarshalling JSON response: ", err.Error()))
	}
	if response.Ok == false {
		return nil, fmt.Errorf(fmt.Sprintf("Error: API response not ok."))
		time.Sleep(time.Minute)
	}
	return &response, nil
}

func sendMessage(chatId int, text string, replyTo int) (*http.Response, error) {
	s := fmt.Sprintf("%s/sendMessage?chat_id=%d&text=%s&reply_to_message_id=%d", Url, chatId, text, replyTo)
	reqS, err := http.NewRequest("POST", s, nil)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error building request: ", err.Error()))
	}
	res, err := Client.Do(reqS)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Error sending request: ", err.Error()))
	}
	return res, nil
}
