package main

import (
	"log"
	"net/http"
	"os"
    "io/ioutil"
    "encoding/json"
    "fmt"
    "time"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

var subjects *Subjects
var subjectsDataMap map[int]SubjectsData = make(map[int]SubjectsData)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.GET("/api/v2/subjects", func(c *gin.Context) {
		if subjects == nil {
			ch := make(chan *Subjects)
			go getSubjects(ch)
			subjects = <-ch
			for i := 0; i<len(subjects.Data); i++ {
				subjectsDataMap[subjects.Data[i].ID] = subjects.Data[i]
			}
		}

		fmt.Printf("%-v\n", subjectsDataMap[19])
		fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
		fmt.Printf("data has length %d\n", len(subjects.Data))
		c.JSON(200, subjects)
	})

	router.GET("/srs/status", func(c *gin.Context) {
		// get assignments
		// get review statistics

		// iterate review statistics
		// - exclude burned_at assignments
		// - exclude not-yet-passed assignments
		// - calculate scores
		// - calculate worst score
		// - remove worst score < 1.0
		// - determine readings and meanings

		// if subjects == nil {
		// 	subjects = getSubjects()
		// 	for i := 0; i<len(subjects.Data); i++ {
		// 		subjectsDataMap[subjects.Data[i].ID] = subjects.Data[i]
		// 	}
		// }

		// fmt.Printf("%-v\n", subjectsDataMap[19])
		// fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
		// fmt.Printf("data has length %d\n", len(subjects.Data))
		c.JSON(200, getReviewStatistics())
	})

	router.Run(":" + port)
}

func getUrl(url string) []byte {
	start := time.Now()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Token token=" + os.Getenv("WANIKANI_V2_API_KEY"))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	secs2 := time.Since(start).Seconds()
  
	fmt.Printf("%f: %s\n", secs2, url)

	return body
}

func getSubjects(chResult chan *Subjects) {
	ch := make(chan *Subjects)

	maxPages := 18
	for page := 1; page <= maxPages; page++ {
		go getSubjectsPage(page, ch)
	}
	
	subjects := <-ch
	if (int(subjects.Pages.Last) > maxPages) {
		for page := maxPages+1; page <= int(subjects.Pages.Last); page++ {
			go getSubjectsPage(page, ch)
		}
		maxPages = int(subjects.Pages.Last)
	}

	for page := 2; page <= maxPages; page++ {
		subjectsPage := <-ch
		subjects.Data = append(subjects.Data, subjectsPage.Data...)
	}

	subjects.Pages.Current = 1

	chResult <- subjects
}

func getSubjectsPage(page int, ch chan *Subjects) {
	body := getUrl(fmt.Sprintf("https://wanikani.com/api/v2/subjects?page=%d",page))
	var subjects Subjects
	
	err := json.Unmarshal(body, &subjects)
	if err != nil {
		log.Fatal("error:", err, string(body))
	}

	ch <- &subjects
}

func getReviewStatistics() *ReviewStatistics {
	ch := make(chan *ReviewStatistics)
	maxPages := 1
	for page := 1; page <= maxPages; page++ {
		go getReviewStatisticsPage(page, ch)
	}
	
	results := <-ch
	if (int(results.Pages.Last) > maxPages) {
		for page := maxPages+1; page <= int(results.Pages.Last); page++ {
			go getReviewStatisticsPage(page, ch)
		}
		maxPages = int(results.Pages.Last)
	}

	for page := 2; page <= maxPages; page++ {
		resultsPage := <-ch
		results.Data = append(results.Data, resultsPage.Data...)
	}

	results.Pages.Current = 1

	return results
}


func getReviewStatisticsPage(page int, ch chan *ReviewStatistics) {
	body := getUrl(fmt.Sprintf("https://wanikani.com/api/v2/review_statistics?page=%d",page))
	var results ReviewStatistics
	
	err := json.Unmarshal(body, &results)
	if err != nil {
		log.Fatal("error:", err, string(body))
	}

	ch <- &results
}