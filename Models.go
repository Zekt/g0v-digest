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
	StartTime int64   `xml:"startYmdUnix"`
	EndTime   int64   `xml:"endYmdUnix"`
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

func StoreArticle(article Article) {
	statement := `
	INSERT INTO article
	(title, lang, pubtime, html, url)
	VALUES ($1, $2, $3, $4, $5)
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
	} else {
		// Renew XML if a new article inserted.
		_, err = DB.Exec("INSERT INTO line_xml (time) VALUES ($1)", time.Now())
		if err != nil {
			log.Println(err.Error())
		}
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
	var line LineXML
	r := DB.QueryRow(`
	SELECT id, (extract(epoch from time)*1000)::bigint
	FROM line_xml ORDER BY id DESC LIMIT 1
	`)
	if err := r.Scan(&line.UUID, &line.Time); err != nil {
		return nil, err
	}

	statement := `
	SELECT id, title, lang, (extract(epoch from pubtime)*1000)::bigint, html, url
	FROM article WHERE pubtime <= to_timestamp(($1)) ORDER BY pubtime DESC
	`
	rs, err := DB.Query(statement, line.Time)
	if err != nil {
		return nil, err
	}

	for rs.Next() {
		var subxml LineArticleXML
		err = rs.Scan(
			&subxml.Id,
			&subxml.Title,
			&subxml.Language,
			&subxml.PubTime,
			&subxml.Html.Html,
			&subxml.Url,
		)
		subxml.Country = "TW"
		subxml.StartTime = 0
		subxml.EndTime = 2000000000000
		subxml.Category = "digest"
		if err != nil {
			log.Println(err.Error())
		}
		line.Articles = append(line.Articles, subxml)
	}

	return xml.Marshal(line)
}
