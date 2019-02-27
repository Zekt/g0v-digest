package main

import (
	//"fmt"
	"io"
	"log"
	//"net/http"

	"github.com/PuerkitoBio/goquery"
)

type SplitedArticle struct {
	Digests []struct{ title, img, content string }
	Remains string
}

func Parse(source io.Reader) SplitedArticle {
	doc, err := goquery.NewDocumentFromReader(source)
	if err != nil {
		log.Println("Error parsing html: ", err.Error())
	}

	var digest SplitedArticle

	nodes := doc.Find("h3")
	nodes.Each(func(index int, node *goquery.Selection) {
		h3, err := node.Html()
		imgSrc := node.Next().Children().AttrOr("src", "")
		p, err := node.Next().Next().Html()
		if err != nil {
			log.Println("Error reading HTML: ", err.Error())
			return
		}
		digest.Digests = append(digest.Digests, struct{ title, img, content string }{h3, imgSrc, p})
		log.Println(node.Next().Next().Html())
	})
	return digest
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
