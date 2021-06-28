package main

import (
	"os"

	"github.com/labstack/echo"
	"github.com/leanrgo/scrapper"
)

const Filename = "articles.csv"

// Handler
func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove("articles.csv")
	scrapper.Scrapper()
	return c.Attachment(Filename, Filename)
}

func main() {
	// scrapper.Scrapper()
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
