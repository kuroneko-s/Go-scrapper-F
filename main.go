package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var baseUrl string = "https://bbs.ruliweb.com/news/board/1001"

// ?page=1
type extractedArticle struct {
	id       string
	division string
	subject  string
	writer   string
	views    string
}

func main() {
	var articles []extractedArticle
	c := make(chan []extractedArticle)
	totalPages := getPages()

	for i := 0; i <= totalPages; i++ {
		go getPage(i, c)
	}
	for i := 0; i < totalPages; i++ {
		extractArticles := <-c
		articles = append(articles, extractArticles...)
	}

	// articles => [{...} ...]
	writeArticles(articles)
	fmt.Println("Done :", len(articles))
}

func getPage(page int, mainC chan<- []extractedArticle) {
	var articles []extractedArticle
	c := make(chan extractedArticle)
	pageURL := baseUrl + "?page=" + strconv.Itoa(page)
	fmt.Println("Requesting ", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".table_body")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractArticle(card, c)
	})
	// get value in chaner

	for i := 0; i < searchCards.Length(); i++ {
		article := <-c
		articles = append(articles, article)
	}
	mainC <- articles
	// return articles
}

func extractArticle(card *goquery.Selection, c chan<- extractedArticle) {
	id := cleanString(card.Find(".id").Text())
	subject := cleanString(card.Find(".subject > .relative > a").Text())
	division := cleanString(card.Find(".divsn > a").Text())
	writer := cleanString(card.Find(".writer > a").Text())
	views := strings.TrimSpace(card.Find(".recomd").Text())
	c <- extractedArticle{id: id, subject: subject, division: division, writer: writer, views: views}
	// return article
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".paging_wrapper").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length() - 1
	})

	return pages
}

// saved file csv
func writeArticles(articles []extractedArticle) {
	file, err := os.Create("./articles.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "DIVISION", "SUBJECT", "WRITER", "VIEWS"}
	wErr := w.Write(headers)
	checkErr(wErr)

	// article => { ... }

	c := make(chan []string)

	for _, article := range articles {
		go writeArticleDetails(article, c)
	}

	for i := 0; i < len(articles); i++ {
		err := w.Write(<-c)
		checkErr(err)
	}
}

func writeArticleDetails(article extractedArticle, c chan<- []string) {
	c <- []string{article.id, article.division, article.subject, article.writer, article.views}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status: ", res.StatusCode)
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
