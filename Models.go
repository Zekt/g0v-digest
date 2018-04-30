package main

import (
	"database/sql"
	"encoding/xml"
	_ "github.com/lib/pq"
	"log"
	"time"
)

var DB *sql.DB

type Article struct {
	Title    string
	Language string
	PubTime  time.Time
	Url      string
	Tags     []string
	Html     string
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
	StartTime int     `xml:"startYmdUnix"`
	EndTime   int     `xml:"endYmdUnix"`
	Title     string  `xml:"title"`
	Category  string  `xml:"category"`
	PubTime   int     `xml:"publishTimeUnix"`
	Html      Content `xml:"contents>text>content"`
	//Html      string `xml:"contents>text>content"`
	Url string `xml:"sourceUrl"`
}

type Content struct {
	XMLName xml.Name `xml:"content"`
	Html    string   `xml:",cdata"`
}

func StoreArticle(article Article) {
	statement := `
	INSERT INTO article
	(title, lang, pubtime, html, url)
	VALUES ($1, $2, $3, $4, $5) ON CONFLICT (url) DO NOTHING
	`
	_, err := DB.Exec(statement,
		article.Title,
		article.Language,
		article.PubTime,
		article.Html,
		article.Url,
	)
	if err != nil {
		log.Println(err.Error())
	}

	r := DB.QueryRow("SELECT id FROM article WHERE title=$1", article.Title)
	var articleId int
	if err := r.Scan(&articleId); err != nil {
		log.Println(err.Error())
	}

	for _, v := range article.Tags {
		statement = `
		INSERT INTO tag (name) VALUES ($1) ON CONFLICT (name) DO NOTHING
		`
		if _, err := DB.Exec(statement, v); err != nil {
			log.Println(err.Error())
		}

		r := DB.QueryRow("SELECT id FROM tag WHERE name=$1", v)

		var tagId int
		if err := r.Scan(&tagId); err != nil {
			log.Println(err.Error())
		} else {
			statement = `
			INSERT INTO map_tag_article
			VALUES ($1, $2) ON CONFLICT DO NOTHING
			`
			if _, err := DB.Exec(statement, tagId, articleId); err != nil {
				log.Println(err.Error())
			}
		}
	}
}

func GetNewestXML() ([]byte, error) {
	statement := `
	SELECT id, title, lang, extract(epoch from pubtime) :: bigint, html, url
	FROM article ORDER BY pubtime DESC
	`
	var line LineArticleXML
	r := DB.QueryRow(statement)
	err := r.Scan(&line.Id, &line.Title, &line.Language, &line.PubTime, &line.Html.Html, &line.Url)
	if err != nil {
		return nil, err
	}

	resObj := &LineXML{
		UUID: "",
		Time: time.Now().Unix(),
		Articles: []LineArticleXML{
			line,
		},
	}

	return xml.Marshal(resObj)
}
