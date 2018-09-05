package main

import (
	"encoding/xml"
	"time"
)

var config struct {
	Server  string `json:"server"`
	Port    int    `json:"port"`
	SQLHost string `json:"sqlHost"`
	SQLPort int    `json:"sqlPort"`
	SQLUser string `json:"sqlUser"`
	SQLPass string `json:"sqlPass"`
	DBName  string `json:"dbname"`
	RssUrl  string `json:"rssUrl"`
	CampId  string `json:"campaignId"`
	TempId  int    `json:"templateId"`
	ListId  string `json:"listId"`
	ApiUrl  string `json:"mailchimpUrl"`
	ApiKey  string `json:"apiKey"`
}

type Article struct {
	Title    string
	Language string
	PubTime  time.Time
	Url      string
	Tags     []string
	Html     string
}

type SplitedArticle struct {
	Digests []struct{ title, pos, img, content string }
	Remains string
}

type Campaign struct {
	Type   string `json:"type"`
	ListId string `json:"recipients>list_id"`
}

type LineXML struct {
	XMLName  xml.Name         `xml:"articles"`
	UUID     string           `xml:"UUID"`
	Time     int64            `xml:"time"`
	Articles []LineArticleXML `xml:"article"`
}

type LineArticleXML struct {
	Id        string  `xml:"ID"`
	Country   string  `xml:"nativeCountry"`
	Language  string  `xml:"language"`
	StartTime int64   `xml:"startYmdtUnix"`
	EndTime   int64   `xml:"endYmdtUnix"`
	Title     string  `xml:"title"`
	Category  string  `xml:"category"`
	PubTime   int64   `xml:"publishTimeUnix"`
	Html      Content `xml:"contents>text>content"`
	//Html      string `xml:"contents>text>content"`
	Url string `xml:"sourceUrl"`
}

type Content struct {
	XMLName xml.Name `xml:"content"`
	Html    string   `xml:",cdata"`
}
