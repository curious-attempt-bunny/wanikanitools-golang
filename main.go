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
		subjects := getSubjects()

		fmt.Printf("%d subjects pages in total\n", subjects.Pages.Last)
		fmt.Printf("data has length %d\n", len(subjects.Data))
		c.JSON(200, subjects)
	})

	router.Run(":" + port)
}

func getSubjects() *Subjects {
	ch := make(chan Subjects)
	maxPages := 5
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
		// fmt.Printf("%d/%d: Appending page %d to page %d\n", page, maxPages, subjectsPage.Pages.Current, subjects.Pages.Current)
		subjects.Data = append(subjects.Data, subjectsPage.Data...)
	}

	subjects.Pages.Current = 1

	return &subjects
}

func getUrl(url string) []byte {
	start := time.Now()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Token token=" + os.Getenv("WANIKANI_V2_API_KEY"))
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	// secs1 := time.Since(start).Seconds()
	body, _ := ioutil.ReadAll(resp.Body)
	secs2 := time.Since(start).Seconds()
  
	fmt.Printf("%f: %s\n", secs2, url)

	return body
}

func getSubjectsPage(page int, ch chan Subjects) {
	body := getUrl(fmt.Sprintf("https://wanikani.com/api/v2/subjects?page=%d",page))
	var subjects Subjects
	// start := time.Now()
	
	err := json.Unmarshal(body, &subjects)
	if err != nil {
		log.Fatal("error:", err, string(body))
	}
	// secs := time.Since(start).Seconds()

	// fmt.Printf("%f: JSON %d\n", secs, page)

	// fmt.Printf("Got subject page %d\n", page)	
	ch <- subjects
}