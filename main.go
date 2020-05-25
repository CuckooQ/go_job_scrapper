package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// constants
const baseURL string = "https://kr.indeed.com/jobs?q=python&limit=50"

// functions
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

func getPages() int {
	pages := 0

	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func getPage(page int) string {
	pageURL := baseURL + "&start=" + strconv.Itoa(page*50)

	return pageURL
}

// main
func main() {
	totalpages := getPages()

	for i := 0; i < totalpages; i++ {
		pageURL := getPage(i)
		fmt.Println(pageURL)
	}
}
