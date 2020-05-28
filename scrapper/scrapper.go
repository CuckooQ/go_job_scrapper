package scrapper

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

// structs
type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

// constants
const viewURL string = "https://kr.indeed.com/viewjob?jk="

// functions
func Scrape(term string) {
	baseURL := "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	jobs := []extractedJob{}
	c := make(chan []extractedJob)
	totalpages := getTotalPages(baseURL)

	for i := 0; i < totalpages; i++ {
		go getJobs(i, baseURL, c)
	}

	for i := 0; i < totalpages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)

	fmt.Println("Done")
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

func getTotalPages(url string) int {
	total := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		total = s.Find("a").Length()
	})

	return total
}

func getJobs(page int, url string, mainC chan<- []extractedJob) {
	jobs := []extractedJob{}
	c := make(chan extractedJob)
	pageURL := url + "&start=" + strconv.Itoa(page*50)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".jobsearch-SerpJobCard")

	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := cleanString(card.Find(".title>a").Text())
	location := cleanString(card.Find(".sjcl").Text())
	salary := cleanString(card.Find(".salaryText").Text())
	summary := cleanString(card.Find(".summary").Text())

	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func writeJobs(jobs []extractedJob) {
	c := make(chan []string)
	jobSlice := []string{}
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)

	// input data to file
	defer w.Flush()

	headers := []string{"id", "title", "location", "salary", "summary"}
	err = w.Write(headers)
	checkErr(err)

	for _, job := range jobs {
		go getJobSlice(job, c)
	}

	for i := 0; i < len(jobs); i++ {
		unitJobSlice := <-c
		jobSlice = append(jobSlice, unitJobSlice...)
	}

	err = w.Write(jobSlice)
	checkErr(err)
}

func getJobSlice(job extractedJob, c chan<- []string) {
	c <- []string{viewURL + job.id, job.title, job.location, job.salary, job.summary}
}
