package main

import (
	"database/sql"
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
			INSERT INTO map_tag_article VALUES ($1, $2) ON CONFLICT DO NOTHING
			`
			if _, err := DB.Exec(statement, tagId, articleId); err != nil {
				log.Println(err.Error())
			}
		}
	}
}
